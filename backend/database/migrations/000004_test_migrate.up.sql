-- 测试 migrate 是否能够创建数据库表。
CREATE TABLE `migration_test` (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `message` VARCHAR(255) NOT NULL,
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci
  COMMENT = '数据库迁移测试表';
