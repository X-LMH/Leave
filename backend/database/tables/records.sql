-- 请假记录表：保存渲染请假条所需的数据。
-- leave_type_id、student_id 由应用层维护关联，不建立数据库外键。
CREATE TABLE `records`
(
    `id`              INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '请假记录主键',
    `student_id`      VARCHAR(22)   NOT NULL COMMENT '学号',
    `name`            VARCHAR(64)   NOT NULL COMMENT '姓名快照',
    `leave_type_id`   INT UNSIGNED NOT NULL COMMENT '请假类型ID',
	`leave_type_name` VARCHAR(64)   NOT NULL COMMENT '请假类型名称快照',
	`college`         VARCHAR(64)   NOT NULL COMMENT '学院快照',
	`major`           VARCHAR(64)   NOT NULL COMMENT '专业快照',
	`class_name`      VARCHAR(64)   NOT NULL COMMENT '班级快照',
    `start_time`      DATETIME      NOT NULL COMMENT '请假开始时间',
    `end_time`        DATETIME      NOT NULL COMMENT '请假结束时间',
    `duration`        INT UNSIGNED NOT NULL COMMENT '请假时长（小时）',
    `affected_course` VARCHAR(255)  NOT NULL DEFAULT '' COMMENT '影响课程',
    `is_leave_school` TINYINT(1)  NOT NULL DEFAULT 0 COMMENT '是否离校：1=是，0=否',
    `leave_reason`    TEXT          NOT NULL COMMENT '请假理由',
    `travel_way`      VARCHAR(32)   NOT NULL DEFAULT '' COMMENT '出行方式',
    `applied_at`      DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '申请时间',
    `approved_at`     DATETIME NULL DEFAULT NULL COMMENT '通过时间',
    `created_at`      DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
    `updated_at`      DATETIME      NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '记录更新时间',
    `deleted_at`      DATETIME NULL DEFAULT NULL COMMENT '软删除时间，NULL表示未删除',
    PRIMARY KEY (`id`),
    KEY               `idx_records_student_id` (`student_id`),
    KEY               `idx_records_applied_at` (`applied_at`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_unicode_ci COMMENT ='请假记录表';
