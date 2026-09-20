ALTER TABLE `records`
    ADD COLUMN `teacher_name` VARCHAR(64) NULL COMMENT '班主任姓名快照' AFTER `parent_phone`;

UPDATE `records` AS r
JOIN `profiles` AS p ON p.`student_id` = r.`student_id`
SET r.`teacher_name` = COALESCE(p.`teacher_name`, '')
WHERE r.`teacher_name` IS NULL OR r.`teacher_name` = '';

ALTER TABLE `records`
    MODIFY COLUMN `teacher_name` VARCHAR(64) NOT NULL COMMENT '班主任姓名快照';
