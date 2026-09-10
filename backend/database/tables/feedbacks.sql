-- 用户反馈表：保存用户提交的文字反馈，匿名反馈不保留提交者学号。
CREATE TABLE `feedbacks`
(
    `id`           INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '反馈主键',
    `student_id`   VARCHAR(22)  NULL COMMENT '提交者学号，匿名反馈为空',
    `is_anonymous` TINYINT(1)   NOT NULL DEFAULT 0 COMMENT '是否匿名：1=是，0=否',
    `content`      TEXT         NOT NULL COMMENT '反馈内容',
    `created_at`   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '提交时间',
    `updated_at`   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at`   DATETIME     NULL     DEFAULT NULL COMMENT '软删除时间，NULL表示未删除',
    PRIMARY KEY (`id`),
    KEY `idx_feedbacks_student_id` (`student_id`),
    KEY `idx_feedbacks_created_at` (`created_at`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci COMMENT ='用户反馈表';
