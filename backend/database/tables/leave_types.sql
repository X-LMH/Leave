CREATE TABLE `leave_types`
(
    `id`         INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `name`       VARCHAR(32)  NOT NULL COMMENT '请假类型名称',
    `sort_order` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '排序值，越小越靠前',
    `is_enabled` TINYINT(1)   NOT NULL DEFAULT 1 COMMENT '是否启用：1=启用，0=停用',
    `created_at` DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at` DATETIME     NULL     DEFAULT NULL COMMENT '软删除时间，NULL表示未删除',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_leave_types_name` (`name`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci COMMENT ='请假类型字典表';
