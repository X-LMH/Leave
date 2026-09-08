-- 应用版本初始化数据
INSERT INTO `app_versions`
    (`platform`, `version_code`, `version_name`, `package_file`, `release_notes`, `status`)
VALUES ('android', 2, '0.0.2', 'leave-android-2-0.0.2.apk', JSON_ARRAY('正式发行版'), 'published');
