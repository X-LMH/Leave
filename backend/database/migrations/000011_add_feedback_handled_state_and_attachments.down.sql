-- 回滚会丢失处理状态和附件路径记录，重新升级无法恢复；需要保留时应先备份。
-- 不会删除上传目录中的附件文件。
ALTER TABLE `feedbacks`
    DROP COLUMN `is_handled`,
    DROP COLUMN `attachments`;
