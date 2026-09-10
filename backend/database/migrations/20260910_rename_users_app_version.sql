-- 将用户使用状态字段调整为最近进入软件状态。
ALTER TABLE `users`
    RENAME COLUMN `last_login_at` TO `last_seen_at`,
    RENAME COLUMN `last_login_device` TO `last_seen_device`,
    RENAME COLUMN `last_login_app_version` TO `app_version`;
