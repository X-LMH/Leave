ALTER TABLE `app_versions`
    MODIFY COLUMN `platform` VARCHAR(20) NOT NULL COMMENT '客户端平台，例如 android';
