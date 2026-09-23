-- 班级归属表：记录学院、专业和班级名称。
CREATE TABLE `classes`
(
    `id`         INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '班级ID',
    `college`    VARCHAR(100) NOT NULL COMMENT '学院',
    `major`      VARCHAR(100) NOT NULL COMMENT '专业',
    `class_name` VARCHAR(64)  NOT NULL COMMENT '班级名称',
    `is_enabled` TINYINT(1)   NOT NULL DEFAULT 1 COMMENT '是否启用：1=启用，0=停用',
    `created_at` DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_classes_full_name` (`college`, `major`, `class_name`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci COMMENT ='班级归属表';
