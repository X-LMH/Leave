-- 用户个人信息表：保留原有字段，仅明确字段类型。
CREATE TABLE `profiles`
(
    `id`           INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '个人信息主键',
    `student_id`   VARCHAR(22)  NOT NULL COMMENT '关联的学号',
    `class_id`     INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '班级ID',
    `name`         VARCHAR(64)  NOT NULL COMMENT '学生姓名',
    `phone`        VARCHAR(20)  NOT NULL COMMENT '学生电话',
    `gender`       VARCHAR(10)  NOT NULL COMMENT '性别：男/女',
    `parent_name`  VARCHAR(64)  NOT NULL COMMENT '家长姓名',
    `parent_phone` VARCHAR(20)  NOT NULL COMMENT '家长电话',
    `apartment_id`      INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '关联的公寓ID',
    `dormitory_number`  VARCHAR(32)  NOT NULL DEFAULT '' COMMENT '宿舍号',
    `teacher_name`      VARCHAR(64)  NULL     DEFAULT NULL COMMENT '辅导员姓名',
    `avatar_url`   VARCHAR(512) NOT NULL DEFAULT '' COMMENT '头像相对路径（相对于上传文件根目录，如 avatars/abc123.jpg）',
    `created_at`    DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at`   DATETIME     NULL     DEFAULT NULL COMMENT '软删除时间，NULL表示未删除',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_profiles_student_id` (`student_id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci COMMENT ='用户个人信息表';
