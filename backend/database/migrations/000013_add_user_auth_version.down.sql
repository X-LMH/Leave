-- 回滚将丢失认证版本信息；重新升级前须清空对应 namespace 的会话，避免旧会话恢复有效。
ALTER TABLE `users` DROP COLUMN `auth_version`;
