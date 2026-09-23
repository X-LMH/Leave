ALTER TABLE `classes`
    ADD COLUMN `is_enabled` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否启用：1=启用，0=停用' AFTER `class_name`;
