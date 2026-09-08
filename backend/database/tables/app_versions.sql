CREATE TABLE `app_versions`
(
    `id`            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `platform`      VARCHAR(20)     NOT NULL COMMENT '客户端平台，例如 android',
    `version_code`  INT UNSIGNED    NOT NULL COMMENT '构建版本号，用于版本校验',
    `version_name`  VARCHAR(32)     NOT NULL COMMENT '展示版本号',
    `download_url`  VARCHAR(500)    NOT NULL COMMENT '安装包下载地址',
    `apk_sha256`    CHAR(64)        NOT NULL DEFAULT '' COMMENT 'APK SHA-256 校验值',
    `release_notes` JSON            NULL COMMENT '更新说明 JSON 数组',
    `status`        VARCHAR(20)     NOT NULL DEFAULT 'published' COMMENT 'published=当前可用，archived=历史版本',
    `published_at`  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '发布时间',
    `created_at`    DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`    DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_app_versions_platform_code` (`platform`, `version_code`),
    KEY `idx_app_versions_current` (`platform`, `status`, `version_code`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci COMMENT ='客户端版本发布记录';