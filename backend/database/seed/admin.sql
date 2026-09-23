-- 默认管理员账号（首次登录后请立即修改密码）
INSERT INTO `users` (`student_id`, `password`, `role`, `status`)
VALUES ('admin', 'admin123', 'admin', 1)
ON DUPLICATE KEY UPDATE `student_id` = VALUES(`student_id`);
