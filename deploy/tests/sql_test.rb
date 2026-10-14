require 'minitest/autorun'
require 'open3'
require_relative '../release'

class VersionSQLTest < Minitest::Test
  def query(sql)
    out, err, status = Open3.capture3(
      'mysql', '--protocol=tcp', '--host=127.0.0.1', '--user=root',
      '--batch', '--skip-column-names', '--default-character-set=utf8mb4',
      stdin_data: sql
    )
    assert status.success?, err
    out.strip
  end

  def test_first_publish_retry_switch_and_overwrite
    skip 'DEPLOY_TEST_MYSQL=1 requires an isolated local MySQL service' unless ENV['DEPLOY_TEST_MYSQL'] == '1'
    database = "leave_release_test_#{Process.pid}_#{Time.now.to_i}"
    query("CREATE DATABASE #{database} CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;")
    begin
      table = File.read(File.join(Release::ROOT, 'backend/database/tables/app_versions.sql'))
      query("USE #{database};\n#{table}")
      info = {
        'tag' => 'app-test-v1.0.0', 'mode' => 'app', 'backend_version' => 1,
        'app' => {
          'version_name' => '1.0.0', 'version_code' => 11,
          'apk_name' => 'leave-android-11-1.0.0.apk',
          'release_notes' => ["中文 ' \\ 更新"]
        }
      }
      2.times { query("USE #{database};\n#{Release.version_sql(info)}") }
      assert_equal '1', query("SELECT COUNT(*) FROM #{database}.app_versions;")
      assert_equal '1', query("SELECT COUNT(*) FROM #{database}.app_versions WHERE status='published';")
      info['app']['release_notes'] = ['覆盖更新']
      query("USE #{database}; SET sql_mode='NO_BACKSLASH_ESCAPES';\n#{Release.version_sql(info)}")
      assert_equal '["覆盖更新"]', query("SELECT release_notes FROM #{database}.app_versions;")
      info['tag'] = 'app-test-v1.0.1'
      info['app'].merge!('version_name' => '1.0.1', 'version_code' => 12, 'apk_name' => 'leave-android-12-1.0.1.apk')
      query("USE #{database};\n#{Release.version_sql(info)}")
      assert_equal "11\tarchived\n12\tpublished", query("SELECT version_code,status FROM #{database}.app_versions ORDER BY version_code;")
    ensure
      query("DROP DATABASE #{database};")
    end
  end
end
