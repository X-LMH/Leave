-- 历史反馈统一初始化为未处理且无附件，不推断管理员是否曾经处理。
-- 附件仅保存相对路径，文件本身存放在上传目录。
ALTER TABLE `feedbacks`
    ADD COLUMN `attachments` JSON NULL
    COMMENT '附件相对路径数组（相对于上传文件根目录，如 ["feedbacks/abc123.jpg"]），无附件为NULL'
    AFTER `content`,
    ADD COLUMN `is_handled` TINYINT(1) NOT NULL DEFAULT 0
    COMMENT '是否已处理：0=未处理，1=已处理；已处理不代表已采纳或已解决'
    AFTER `attachments`;
