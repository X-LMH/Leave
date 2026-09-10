-- 用户账号表：只保存登录身份、认证信息和账号状态。
CREATE TABLE `users`
(
    `id`            INT UNSIGNED     NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `student_id`    VARCHAR(22)      NOT NULL COMMENT '学号',
    `password` VARCHAR(255)     NOT NULL COMMENT '密码',
    `role`          VARCHAR(20)      NOT NULL DEFAULT 'student' COMMENT '角色：student=学生，admin=管理员',
    `status`        TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '账号状态：1=正常，0=停用',
    `last_seen_at` DATETIME         NULL     DEFAULT NULL COMMENT '最近进入软件时间',
    `last_seen_device` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '最近进入软件设备名称',
    `app_version`           VARCHAR(64) NOT NULL DEFAULT '' COMMENT '当前应用版本',
    `created_at`    DATETIME         NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`    DATETIME         NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at`    DATETIME         NULL     DEFAULT NULL COMMENT '软删除时间，NULL表示未删除',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_users_student_id` (`student_id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci COMMENT ='用户账号表';
