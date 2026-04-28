-- Auto-generated schema for casbin_model
-- Database: goadmin
CREATE TABLE IF NOT EXISTS `casbin_model` (
  `name` varchar(64) NOT NULL,
  `content` varchar(255) NOT NULL,
  `created_at` datetime(3) NULL,
  `updated_at` datetime(3) NULL,
  PRIMARY KEY (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
