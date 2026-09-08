-- 应用版本初始化数据
INSERT INTO `app_versions`
    (`platform`, `version_code`, `version_name`, `download_url`, `apk_sha256`, `release_notes`, `status`)
VALUES ('android', 1, '0.0.1', 'http://www.example.com', '', JSON_ARRAY('测试版本'), 'published');
