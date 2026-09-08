-- 应用版本初始化数据
INSERT INTO `app_versions`
    (`platform`, `version_code`, `version_name`, `package_file`, `release_notes`, `status`)
VALUES ('android', 1, '0.0.1', 'leave-android-1-0.0.1.apk', JSON_ARRAY('测试版本'), 'published');
