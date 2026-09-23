ALTER TABLE `records`
    ADD COLUMN `gender` VARCHAR(10) NOT NULL DEFAULT '' COMMENT '性别快照：male/female' AFTER `name`;

UPDATE `records` AS r
    INNER JOIN `profiles` AS p ON p.`student_id` = r.`student_id`
SET r.`gender` = p.`gender`
WHERE r.`gender` = '';
