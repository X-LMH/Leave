-- 公寓信息表：维护可供学生选择的男生、女生公寓。
create table `apartments`
(
    `id`         int unsigned     not null auto_increment comment '公寓ID',
    `name`       varchar(100)     not null comment '公寓名称',
    `gender`     varchar(10)      not null comment '性别',
    `sort_order` int unsigned     not null default 0 comment '排序值，数值越小越靠前',
    `is_enabled` tinyint unsigned not null default 1 comment '是否启用：1=启用，0=停用',
    `created_at` datetime         not null default current_timestamp comment '创建时间',
    `updated_at` datetime         not null default current_timestamp on update current_timestamp comment '更新时间',
    primary key (`id`),
    unique key `uk_apartments_name_gender` (`name`, `gender`),
    key `idx_apartments_enabled_sort_order` (`is_enabled`, `sort_order`)
) engine = innodb
  default charset = utf8mb4
  collate = utf8mb4_unicode_ci comment ='公寓信息表';
