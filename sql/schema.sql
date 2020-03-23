-- 列名不能使用双引号括起来
-- unix时间戳从1开始，即1970-01-01 00:00:01，0是非法的unix时间戳
CREATE TABLE IF NOT EXISTS `users` (
	`id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
	`resource_id` varchar(20) NOT NULL COMMENT '资源ID',
	`name` varchar(36) NOT NULL COMMENT '资源名称',
	`deleted_at` timestamp NOT NULL DEFAULT '1970-01-01 00:00:01' COMMENT '删除时间',
	`age` int NOT NULL COMMENT '用户年龄',
	`desc` varchar(255) NOT NULL COMMENT '用户描述',
	PRIMARY KEY (`id`),
	UNIQUE KEY `idx_resource` (`resource_id`, `name`, `deleted_at`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8 AUTO_INCREMENT = 1 COMMENT ='用户表';