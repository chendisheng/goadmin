-- Auto-generated schema for casbin_rules
-- Database: goadmin
CREATE TABLE IF NOT EXISTS `casbin_rules` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `ptype` varchar(255) NOT NULL,
  `v0` varchar(255) NOT NULL,
  `v1` varchar(255) NOT NULL,
  `v2` varchar(255) NOT NULL,
  `v3` varchar(255) NOT NULL,
  `v4` varchar(255) NOT NULL,
  `v5` varchar(255) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
