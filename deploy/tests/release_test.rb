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

  def test_invalid_tags
    %w[v1 back-prod-v1 back-v0 admin-v02 back-v1.0.0 app-v1 back-test-v0 back-test-v01 back-test-v1.0.0 app-test-v1].each do |tag|
      assert_raises(StandardError, tag) { Release.read(tag, @root) }
    end
  end

  def test_production_tags_and_environment_integrity
    prepare_app
    %w[back-v1 admin-v2 app-v1.0.0].each do |tag|
      info = Release.read(tag, @root)
      assert_equal 'prod', Release.environment(info)
      assert_equal info, Release.validate(info)
      info['environment'] = 'test'
      assert_raises(StandardError) { Release.validate(info) }
    end
    assert_equal 'test', Release.environment(Release.read('back-test-v1', @root))
    # Existing test releases have no explicit environment field.
    assert_equal 'test', Release.environment({ 'tag' => 'back-test-v1' })
    assert_raises(StandardError) { Release.read('admin-v3', @root) }
  end

  def test_server_config_rejects_cross_environment_database_and_storage
    %w[test prod].each do |env|
      profile = Release::ENVIRONMENTS.fetch(env)
      config = {
        'mysql' => { 'database' => profile['database'], 'host' => '127.0.0.1' },
        'app' => { 'port' => 8080, 'package_dir' => "#{profile['root']}/packages/android" },
        'storage' => { 'root_dir' => "#{profile['root']}/uploads" }
      }
      file = File.join(@root, 'config.yaml')
      File.write(file, YAML.dump(config))
      assert_equal config, Release.server_config(file, env)
      assert_raises(StandardError) { Release.server_config(file, env == 'test' ? 'prod' : 'test') }
      config['app']['package_dir'] = '/www/wwwroot/other/packages/android'
      File.write(file, YAML.dump(config))
      assert_raises(StandardError) { Release.server_config(file, env) }
      config['app']['package_dir'] = "#{profile['root']}/packages/android"
      config['storage']['root_dir'] = '/www/wwwroot/other/uploads'
      File.write(file, YAML.dump(config))
      assert_raises(StandardError) { Release.server_config(file, env) }
    end
  end

  def test_admin_bundle_uses_selected_environment_build
    %w[bin backend/database/migrations admin/dist admin/dist-test].each { |d| FileUtils.mkdir_p(File.join(@root, d)) }
    File.write(File.join(@root, 'bin/Leave'), 'server')
    File.write(File.join(@root, 'bin/migrate'), 'migration tool')
    File.write(File.join(@root, 'backend/database/migrations/000001_test.up.sql'), 'SELECT 1;')
    %w[release.rb deploy-release.sh switch-and-restart.sh run-migrations.sh publish-version.sh].each do |n|
      FileUtils.cp(File.join(Release::ROOT, 'deploy', n), File.join(@root, 'deploy', n))
    end
    File.write(File.join(@root, 'admin/dist/index.html'), 'production page')
    File.write(File.join(@root, 'admin/dist-test/index.html'), 'test page')
    { 'admin-v2' => 'production page', 'admin-test-v2' => 'test page' }.each do |tag, page|
      bundle = File.join(@root, tag)
      info = Release.read(tag, @root)
      Release.bundle(info, bundle, @root)
      assert_equal info['environment'], Release.environment(Release.verify_bundle(bundle))
      assert_equal page, File.read(File.join(bundle, 'admin/index.html'))
    end
  end

  def test_version_mismatch
    prepare_app
    cases = [
      ['back-test-v2', 'backend.version=1', '2'],
      ['admin-test-v3', 'admin.version=2', '3'],
      ['app-test-v1.0.1', 'versionName="1.0.0"', '1.0.1']
    ]
    script = File.expand_path('../read-release.rb', __dir__)
    cases.each do |tag, field, version|
      error = assert_raises(StandardError) { Release.read(tag, @root) }
      assert_includes error.message, tag
      assert_includes error.message, field
      assert_includes error.message, "标签版本 #{version}"
      assert_includes error.message, '发布已停止'
      stdout, stderr, status = Open3.capture3('ruby', script, '--tag', tag, '--root', @root, '--format', 'github')
      assert_equal 1, status.exitstatus
      assert_empty stdout
      assert_includes stderr, error.message
    end
  end

  def test_admin_and_app_versions_are_independent_of_backend
    prepare_app
    admin = Release.read('admin-test-v2', @root)
    assert_equal 1, admin['backend_version']
    assert_equal 2, admin['admin_version']
    assert_equal admin, Release.validate(admin)
    app = Release.read('app-test-v1.0.0', @root)
    assert_equal 1, app['backend_version']
    assert_equal '1.0.0', app['app']['version_name']
    assert_equal app, Release.validate(app)
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
