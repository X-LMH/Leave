require 'minitest/autorun'
require 'tmpdir'
require 'open3'
require_relative '../release'

class ReleaseTest < Minitest::Test
  def setup
    @root = Dir.mktmpdir('leave-release-')
    FileUtils.mkdir_p(File.join(@root, 'deploy/releases'))
    write_config
  end

  def teardown
    FileUtils.remove_entry(@root)
  end

  def write_config(extra = {})
    config = {
      'backend' => { 'version' => 1 },
      'admin' => { 'version' => 2 },
      'app' => { 'android' => { 'apk_directory' => 'deploy/releases', 'release_notes' => ["中文 ' \\ 更新"] } }
    }.merge(extra)
    File.write(File.join(@root, 'deploy/release.yaml'), YAML.dump(config))
  end

  def prepare_app(count = 1)
    FileUtils.mkdir_p(File.join(@root, 'frontend'))
    File.write(File.join(@root, 'frontend/manifest.json'), '{"versionName":"1.0.0","versionCode":"11"}')
    count.times { |i| File.write(File.join(@root, "deploy/releases/package #{i}.apk"), 'apk') }
  end

  def test_non_app_does_not_require_manifest_or_apk
    write_config('app' => nil)
    assert_equal 'back', Release.read('back-test-v1', @root)['mode']
    assert_equal 2, Release.read('admin-test-v2', @root)['admin_version']
  end

  def test_invalid_and_production_tags
    %w[v1 admin-v2 app-v1.0.0 back-test-v0 back-test-v01 back-test-v1.0.0 admin-test-v1 app-test-v1].each do |tag|
      assert_raises(StandardError, tag) { Release.read(tag, @root) }
    end
  end

  def test_version_mismatch
    assert_raises(StandardError) { Release.read('back-test-v2', @root) }
    prepare_app
    assert_raises(StandardError) { Release.read('app-test-v1.0.1', @root) }
  end

  def test_apk_count_and_arbitrary_filename
    prepare_app(0)
    assert_raises(StandardError) { Release.read('app-test-v1.0.0', @root) }
    File.write(File.join(@root, 'deploy/releases/任意名称.apk'), 'apk')
    info = Release.read('app-test-v1.0.0', @root)
    assert_equal 'leave-android-11-1.0.0.apk', info['app']['apk_name']
    File.write(File.join(@root, 'deploy/releases/extra.apk'), 'apk')
    assert_raises(StandardError) { Release.read('app-test-v1.0.0', @root) }
  end

  def test_apk_version_mismatch
    analyzer = File.join(@root, 'analyzer')
    File.write(analyzer, "#!/bin/sh\nif [ \"$2\" = version-name ]; then echo 1.0.0; else echo 11; fi\n")
    File.chmod(0755, analyzer)
    script = File.expand_path('../verify-apk.sh', __dir__)
    _, _, status = Open3.capture3({ 'APKANALYZER' => analyzer }, 'bash', script, 'anything.apk', '1.0.0', '11')
    assert status.success?
    _, _, status = Open3.capture3({ 'APKANALYZER' => analyzer }, 'bash', script, 'anything.apk', '1.0.1', '11')
    refute status.success?
  end

  def test_sql_supports_retry_and_mode_independent_strings
    prepare_app
    sql = Release.version_sql(Release.read('app-test-v1.0.0', @root))
    assert_includes sql, 'ON DUPLICATE KEY UPDATE'
    assert_includes sql, 'CONVERT(0x'
    refute_includes sql, "中文 '"
    assert_raises(StandardError) { Release.version_sql(Release.read('back-test-v1', @root)) }
  end

  def test_bundle_integrity_and_symlinks
    %w[bin backend/database/migrations].each { |d| FileUtils.mkdir_p(File.join(@root, d)) }
    File.write(File.join(@root, 'bin/Leave'), 'server')
    File.write(File.join(@root, 'bin/migrate'), 'migration tool')
    File.write(File.join(@root, 'backend/database/migrations/000001_test.up.sql'), 'SELECT 1;')
    %w[release.rb deploy-release.sh switch-and-restart.sh run-migrations.sh publish-version.sh].each do |n|
      FileUtils.cp(File.join(Release::ROOT, 'deploy', n), File.join(@root, 'deploy', n))
    end
    bundle = File.join(@root, 'bundle')
    Release.bundle(Release.read('back-test-v1', @root), bundle, @root)
    assert_equal 'back', Release.verify_bundle(bundle)['mode']
    File.write(File.join(bundle, 'backend/Leave'), 'corrupted')
    assert_raises(StandardError) { Release.verify_bundle(bundle) }
    File.write(File.join(bundle, 'backend/Leave'), 'server')
    File.symlink('/etc/passwd', File.join(bundle, 'leak'))
    assert_raises(StandardError) { Release.verify_bundle(bundle) }
  end
end
