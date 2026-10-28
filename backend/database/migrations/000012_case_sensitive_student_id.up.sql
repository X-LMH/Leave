-- 仅调整账号及关联字段，其他字段保持不变。
ALTER TABLE `users`
    MODIFY COLUMN `student_id` VARCHAR(22) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin
    NOT NULL COMMENT '学号';

ALTER TABLE `profiles`
    MODIFY COLUMN `student_id` VARCHAR(22) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin
    NOT NULL COMMENT '关联的学号';

ALTER TABLE `records`
    MODIFY COLUMN `student_id` VARCHAR(22) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin
    NOT NULL COMMENT '学号';

ALTER TABLE `feedbacks`
    MODIFY COLUMN `student_id` VARCHAR(22) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin
    NULL COMMENT '提交者学号，匿名反馈为空';
