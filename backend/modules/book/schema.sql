-- Auto-generated schema for books
CREATE TABLE IF NOT EXISTS `books` (
  `id` varchar(64) NOT NULL,
  `tenant_id` varchar(255) NOT NULL,
  `title` varchar(255) NOT NULL,
  `author` varchar(255) NOT NULL,
  `isbn` varchar(255) NOT NULL,
  `publisher` varchar(255) NOT NULL,
  `publish_date` datetime NULL,
  `category` varchar(255) NOT NULL,
  `description` varchar(255) NOT NULL,
  `status` varchar(255) NOT NULL,
  `price` bigint NOT NULL DEFAULT 0,
  `stock_quantity` bigint NOT NULL DEFAULT 0,
  `cover_image_url` varchar(255) NOT NULL,
  `tags` varchar(255) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
