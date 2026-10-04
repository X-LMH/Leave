ALTER TABLE `profiles`
    ADD COLUMN `avatar_url` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '头像访问地址' AFTER `teacher_name`;
