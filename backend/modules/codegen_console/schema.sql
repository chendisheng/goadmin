-- Auto-generated schema for codegen_consoles
-- Database: goadmin
CREATE TABLE IF NOT EXISTS `codegen_consoles` (
  `id` varchar(64) NOT NULL,
  `name` varchar(255) NOT NULL,
  `enabled` bigint NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
