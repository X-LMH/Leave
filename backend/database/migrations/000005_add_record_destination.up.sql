ALTER TABLE `records`
    ADD COLUMN `destination` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '请假去向' AFTER `travel_way`;

UPDATE `records`
SET `destination` = '内蒙古自治区,呼和浩特市,土默特左旗';
