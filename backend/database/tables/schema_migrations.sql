-- 迁移执行记录表：记录 tables 目录中的表结构脚本是否已经执行。
CREATE TABLE `schema_migrations`
(
    `version` BIGINT     NOT NULL COMMENT '当前迁移版本',
    `dirty`   TINYINT(1) NOT NULL COMMENT '是否处于未完成状态：1=是，0=否',
    PRIMARY KEY (`version`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci COMMENT ='数据库迁移执行记录表';
