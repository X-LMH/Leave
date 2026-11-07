-- Only changes metadata; rollback does not convert categorized data to arrays.
ALTER TABLE `app_versions` MODIFY COLUMN `release_notes` JSON NULL COMMENT '更新说明 JSON 数组';
