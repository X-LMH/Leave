-- 仅调整账号及关联字段，其他字段保持不变。
-- 回滚前备份并处理旧规则下等价的账号（包括大小写、重音差异）。
-- admin / Admin 等并存时唯一索引会拒绝回滚，不得直接合并或删除业务数据。
-- 多表 DDL 非原子操作，执行前逐表检查冲突并在维护窗口操作。
ALTER TABLE `users`
    MODIFY COLUMN `student_id` VARCHAR(22) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci
    NOT NULL COMMENT '学号';

ALTER TABLE `profiles`
    MODIFY COLUMN `student_id` VARCHAR(22) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci
    NOT NULL COMMENT '关联的学号';

ALTER TABLE `records`
    MODIFY COLUMN `student_id` VARCHAR(22) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci
    NOT NULL COMMENT '学号';

ALTER TABLE `feedbacks`
    MODIFY COLUMN `student_id` VARCHAR(22) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci
    NULL COMMENT '提交者学号，匿名反馈为空';
