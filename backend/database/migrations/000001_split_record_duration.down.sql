ALTER TABLE `records`
    ADD COLUMN `duration` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '请假时长（小时）' AFTER `end_time`;

UPDATE `records`
SET `duration` = CEILING(TIMESTAMPDIFF(SECOND, `start_time`, `end_time`) / 3600.0);
