ALTER TABLE `users` ADD COLUMN `auth_version` BIGINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '认证版本，递增后旧会话失效' AFTER `status`;
