ALTER TABLE `profiles`
    MODIFY COLUMN `avatar_url` VARCHAR(512) NOT NULL DEFAULT '' COMMENT '头像相对路径（相对于上传文件根目录，如 avatars/abc123.jpg）';
