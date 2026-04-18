-- Auto-generated schema for casbin_models
-- Database: goadmin
CREATE TABLE IF NOT EXISTS `casbin_models` (
  `name` varchar(64) NOT NULL,
  `content` varchar(255) NOT NULL,
  PRIMARY KEY (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
