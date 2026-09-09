-- Leave 数据库初始化脚本
-- 数据库类型：MySQL 8+
-- 执行顺序：建表 -> 初始化基础数据

-- ==================== 建表 ====================

-- 用户账号表
CREATE TABLE `users`
(
    `id`                    INT UNSIGNED     NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `student_id`            VARCHAR(22)      NOT NULL COMMENT '学号',
    `password`              VARCHAR(255)     NOT NULL COMMENT '密码',
    `role`                  VARCHAR(20)      NOT NULL DEFAULT 'student' COMMENT '角色：student=学生，admin=管理员',
    `status`                TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '账号状态：1=正常，0=停用',
    `last_login_at`         DATETIME         NULL DEFAULT NULL COMMENT '最近登录时间',
    `last_login_device`     VARCHAR(255)     NOT NULL DEFAULT '' COMMENT '最近登录设备名称',
    `last_login_app_version` VARCHAR(64)     NOT NULL DEFAULT '' COMMENT '最近登录应用版本',
    `created_at`            DATETIME         NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`            DATETIME         NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at`            DATETIME         NULL DEFAULT NULL COMMENT '软删除时间，NULL表示未删除',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_users_student_id` (`student_id`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '用户账号表';

-- 班级归属表
CREATE TABLE `classes`
(
    `id`         INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '班级ID',
    `college`    VARCHAR(100) NOT NULL COMMENT '学院',
    `major`      VARCHAR(100) NOT NULL COMMENT '专业',
    `class_name` VARCHAR(64) NOT NULL COMMENT '班级名称',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_classes_full_name` (`college`, `major`, `class_name`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '班级归属表';

-- 公寓信息表
CREATE TABLE `apartments`
(
    `id`         INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '公寓ID',
    `name`       VARCHAR(100) NOT NULL COMMENT '公寓名称',
    `gender`     VARCHAR(10) NOT NULL COMMENT '性别',
    `sort_order` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '排序值，数值越小越靠前',
    `is_enabled` TINYINT UNSIGNED NOT NULL DEFAULT 1 COMMENT '是否启用：1=启用，0=停用',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_apartments_name_gender` (`name`, `gender`),
    KEY `idx_apartments_enabled_sort_order` (`is_enabled`, `sort_order`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '公寓信息表';

-- 用户个人信息表
CREATE TABLE `profiles`
(
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '个人信息主键',
    `student_id` VARCHAR(22) NOT NULL COMMENT '关联的学号',
    `class_id` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '班级ID',
    `name` VARCHAR(64) NOT NULL COMMENT '学生姓名',
    `phone` VARCHAR(20) NOT NULL COMMENT '学生电话',
    `gender` VARCHAR(10) NOT NULL COMMENT '性别：男/女',
    `parent_name` VARCHAR(64) NOT NULL COMMENT '家长姓名',
    `parent_phone` VARCHAR(20) NOT NULL COMMENT '家长电话',
    `apartment_id` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '关联的公寓ID',
    `dormitory_number` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '宿舍号',
    `teacher_name` VARCHAR(64) NULL DEFAULT NULL COMMENT '辅导员姓名',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at` DATETIME NULL DEFAULT NULL COMMENT '软删除时间，NULL表示未删除',
    PRIMARY KEY (`id`), UNIQUE KEY `uk_profiles_student_id` (`student_id`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '用户个人信息表';

-- 请假类型字典表
CREATE TABLE `leave_types`
(
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `name` VARCHAR(32) NOT NULL COMMENT '请假类型名称',
    `sort_order` INT UNSIGNED NOT NULL DEFAULT 0 COMMENT '排序值，越小越靠前',
    `is_enabled` TINYINT(1) NOT NULL DEFAULT 1 COMMENT '是否启用：1=启用，0=停用',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `deleted_at` DATETIME NULL DEFAULT NULL COMMENT '软删除时间，NULL表示未删除',
    PRIMARY KEY (`id`), UNIQUE KEY `uk_leave_types_name` (`name`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '请假类型字典表';

-- 请假记录表
CREATE TABLE `records`
(
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '请假记录主键',
    `student_id` VARCHAR(22) NOT NULL COMMENT '学号',
    `name` VARCHAR(64) NOT NULL COMMENT '姓名快照',
    `parent_name` VARCHAR(64) NOT NULL COMMENT '家长姓名快照',
    `parent_phone` VARCHAR(20) NOT NULL COMMENT '家长电话快照',
    `leave_type_id` INT UNSIGNED NOT NULL COMMENT '请假类型ID',
    `leave_type_name` VARCHAR(64) NOT NULL COMMENT '请假类型名称快照',
    `college` VARCHAR(64) NOT NULL COMMENT '学院快照',
    `major` VARCHAR(64) NOT NULL COMMENT '专业快照',
    `class_name` VARCHAR(64) NOT NULL COMMENT '班级快照',
    `start_time` DATETIME NOT NULL COMMENT '请假开始时间',
    `end_time` DATETIME NOT NULL COMMENT '请假结束时间',
    `duration` INT UNSIGNED NOT NULL COMMENT '请假时长（小时）',
    `affected_course` VARCHAR(255) NOT NULL DEFAULT '' COMMENT '影响课程',
    `is_leave_school` TINYINT(1) NOT NULL DEFAULT 0 COMMENT '是否离校：1=是，0=否',
    `leave_reason` TEXT NOT NULL COMMENT '请假理由',
    `travel_way` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '出行方式',
    `applied_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '申请时间',
    `approved_at` DATETIME NULL DEFAULT NULL COMMENT '通过时间',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '记录创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '记录更新时间',
    `deleted_at` DATETIME NULL DEFAULT NULL COMMENT '软删除时间，NULL表示未删除',
    PRIMARY KEY (`id`), KEY `idx_records_student_id` (`student_id`), KEY `idx_records_applied_at` (`applied_at`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '请假记录表';

-- 客户端版本发布记录
CREATE TABLE `app_versions`
(
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `platform` VARCHAR(20) NOT NULL COMMENT '客户端平台，例如 android',
    `version_code` INT UNSIGNED NOT NULL COMMENT '构建版本号，用于版本校验',
    `version_name` VARCHAR(32) NOT NULL COMMENT '展示版本号',
    `package_file` VARCHAR(255) NOT NULL COMMENT '安装包文件名，仅允许目录内的 APK 文件',
    `release_notes` JSON NULL COMMENT '更新说明 JSON 数组',
    `status` VARCHAR(20) NOT NULL DEFAULT 'published' COMMENT 'published=当前可用，archived=历史版本',
    `published_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '发布时间',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`), UNIQUE KEY `uk_app_versions_platform_code` (`platform`, `version_code`),
    KEY `idx_app_versions_current` (`platform`, `status`, `version_code`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COLLATE = utf8mb4_unicode_ci COMMENT = '客户端版本发布记录';

-- ==================== Seed ====================

INSERT INTO `classes` (`college`, `major`, `class_name`) VALUES
('智能科学与技术学院（网络空间安全学院）', '计算机科学与技术', '计24-1'),
('智能科学与技术学院（网络空间安全学院）', '计算机科学与技术', '计24-2'),
('智能科学与技术学院（网络空间安全学院）', '计算机科学与技术', '计24-3');

INSERT INTO `apartments` (`name`, `gender`, `sort_order`, `is_enabled`) VALUES
('金川2号楼', 'male', 1, 1), ('金川A号楼', 'female', 1, 1);

INSERT INTO `leave_types` (`name`, `sort_order`) VALUES
('事假-本科生', 1), ('病假-本科生', 2);

INSERT INTO `app_versions`
    (`platform`, `version_code`, `version_name`, `package_file`, `release_notes`, `status`)
VALUES
    ('android', 1, '0.0.1', 'leave-android-1-0.0.1.apk', JSON_ARRAY('正式发行版'), 'published');
