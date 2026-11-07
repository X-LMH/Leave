require 'minitest/autorun'
require 'tmpdir'
require_relative '../github-release'

class GitHubReleaseTest < Minitest::Test
  def setup
    @dir = Dir.mktmpdir('release-test-')
    @tag = 'app-v1.2.0'
    @apk = 'leave-android-17-1.2.0.apk'
    File.write(File.join(@dir, @apk), 'apk bytes')
    File.write(File.join(@dir, 'SHA256SUMS'), "#{Digest::SHA256.file(File.join(@dir, @apk)).hexdigest}  #{@apk}\n")
    File.write(File.join(@dir, 'notes.md'), 'Release notes')
    File.write(File.join(@dir, 'metadata.json'), JSON.generate({tag: @tag, title: 'v1.2.0', apk_name: @apk}))
    @calls = []
    @assets = []
    @draft = true
    @releases = []
    @corrupt = false
  end

  def teardown
    FileUtils.remove_entry(@dir)
  end

  def fake_gh(*args)
    @calls << args
    case args[1..2]
    when ['api', '--paginate'] then JSON.generate([@releases])
    when ['release', 'view'] then JSON.generate({isDraft: @draft, assets: @assets.map { |name| {name: name} }, body: 'Release notes'})
    when ['release', 'upload'] then @assets << File.basename(args[4]); ''
    when ['release', 'download']
      name = args[args.index('--pattern') + 1]
      destination = args[args.index('--dir') + 1]
      File.write(File.join(destination, name), @corrupt ? 'different bytes' : File.binread(File.join(@dir, name)))
      ''
    else ''
    end
  end

  def publish
    GitHubRelease.stub(:run, method(:fake_gh)) { GitHubRelease.publish(@dir, 'owner/repo', @tag) }
  end

  def test_first_publish_verifies_assets_before_publication
    assert_includes publish, @tag
    assert @calls.any? { |a| a[1..2] == ['release', 'create'] && a.include?('--draft') && a.include?('--verify-tag') }
    assert_equal 2, @calls.count { |a| a[1..2] == ['release', 'download'] }
    assert_equal ['release', 'edit'], @calls.last[1..2]
    assert_includes @calls.last, '--latest=true'
  end

  def test_partial_draft_retry_only_uploads_missing_asset
    @releases = [{tag_name: @tag, draft: true, prerelease: false}]
    @assets = [@apk]
    publish
    assert_equal 1, @calls.count { |a| a[1..2] == ['release', 'upload'] }
    refute @calls.any? { |a| a[1..2] == ['release', 'create'] }
  end

  def test_published_release_is_reused
    @draft = false
    @releases = [{tag_name: @tag, draft: false, prerelease: false}]
    @assets = [@apk, 'SHA256SUMS']
    publish
    refute @calls.any? { |a| %w[upload edit create].include?(a[2]) }
  end

  def test_conflicting_asset_stops_publication
    @assets = [@apk]
    @corrupt = true
    assert_raises(RuntimeError) { publish }
    refute @calls.any? { |a| a[1..2] == ['release', 'edit'] }
  end

  def test_public_missing_asset_is_not_modified
    @draft = false
    @releases = [{tag_name: @tag, draft: false, prerelease: false}]
    assert_raises(RuntimeError) { publish }
    refute @calls.any? { |a| a[1..2] == ['release', 'upload'] }
  end

  def test_older_version_does_not_become_latest
    @releases = [{tag_name: 'app-v1.3.0', draft: false, prerelease: false}]
    publish
    assert_includes @calls.last, '--latest=false'
  end

  def test_preparation_formats_categories_and_links
    info = {'tag' => @tag, 'app' => {'version_name' => '1.2.0', 'apk_name' => @apk, 'release_notes' => {'features' => ['新增'], 'notices' => ['重新登录']}}}
    body = GitHubRelease.markdown(info, 'owner/repo', 'app-v1.1.0')
    assert_includes body, '### 新增功能'
    assert_includes body, '### 更新须知'
    refute_includes body, '### 问题修复'
    assert_includes body, "releases/download/#{@tag}/#{@apk}"
    assert_includes body, 'compare/app-v1.1.0...app-v1.2.0'
    info['app']['release_notes'] = []
    assert_includes GitHubRelease.markdown(info, 'owner/repo', nil), "commits/#{@tag}"
    assert_includes GitHubRelease.markdown(info, 'owner/repo', nil), '暂无更新说明'
  end

  def test_previous_tag_excludes_newer_test_and_other_module_tags
    responses = {
      ['git', 'tag', '--merged', 'HEAD'] => "app-v1.1.0\napp-v1.0.0\napp-v9.0.0\napp-test-v1.1.5\nback-v8\n",
      ['git', 'rev-list', '-1', 'app-v1.1.0'] => "recent\n",
      ['git', 'rev-list', '-1', 'app-v1.0.0'] => "old\n",
      ['git', 'rev-list', 'HEAD'] => "head\nrecent\nold\n"
    }
    GitHubRelease.stub(:run, ->(*args) { responses.fetch(args) }) do
      assert_equal 'app-v1.1.0', GitHubRelease.previous_tag(@tag)
    end
  end

  def test_normalized_grouped_notes_survive_bundle_verification
    Dir.mktmpdir('release-root-') do |root|
      FileUtils.mkdir_p(File.join(root, 'deploy/releases'))
      FileUtils.mkdir_p(File.join(root, 'frontend'))
      FileUtils.mkdir_p(File.join(root, 'bin'))
      FileUtils.mkdir_p(File.join(root, 'backend/database/migrations'))
      File.write(File.join(root, 'bin/Leave'), 'server')
      File.write(File.join(root, 'bin/migrate'), 'migrate')
      File.write(File.join(root, 'backend/database/migrations/000001_test.up.sql'), 'SELECT 1;')
      %w[release.rb deploy-release.sh switch-and-restart.sh run-migrations.sh publish-version.sh].each do |name|
        FileUtils.cp(File.join(Release::ROOT, 'deploy', name), File.join(root, 'deploy', name))
      end
      File.write(File.join(root, 'deploy/releases/app.apk'), 'apk')
      File.write(File.join(root, 'frontend/manifest.json'), '{"versionName":"1.2.0","versionCode":"17"}')
      File.write(File.join(root, 'deploy/release.yaml'), YAML.dump({'backend' => {'version' => 1}, 'app' => {'android' => {'apk_directory' => 'deploy/releases', 'release_notes' => {'features' => ['新增']}}}}))
      info = Release.read(@tag, root)
      destination = File.join(root, 'bundle')
      Release.bundle(info, destination, root)
      assert_equal ['新增'], Release.verify_bundle(destination).dig('app', 'release_notes', 'features')
      assert_equal [], Release.verify_bundle(destination).dig('app', 'release_notes', 'notices')
    end
  end

  def test_workflow_gates_publication_on_production_app_deploy_success
    workflow = YAML.load_file(File.join(Release::ROOT, '.github/workflows/deploy.yml'))
    job = workflow.fetch('jobs').fetch('release')
    assert_includes job['needs'], 'deploy'
    %w[success app prod].each { |condition| assert_includes job['if'], "'#{condition}'" }
    assert_equal 'write', job.dig('permissions', 'contents')
    assert_equal 'read', workflow.dig('permissions', 'contents')
  end
end
