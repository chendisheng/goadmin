-- Auto-generated schema for book
-- Database: goadmin
CREATE TABLE IF NOT EXISTS `book` (
  `id` varchar(64) NOT NULL COMMENT '主键ID',
  `tenant_id` varchar(255) NOT NULL COMMENT '租户ID',
  `title` varchar(255) NOT NULL COMMENT '书名',
  `author` varchar(255) NOT NULL COMMENT '作者',
  `isbn` varchar(255) NOT NULL COMMENT 'ISBN编号',
  `publisher` varchar(255) NOT NULL COMMENT '出版社',
  `publish_date` datetime NULL COMMENT '出版日期',
  `category` varchar(255) NOT NULL COMMENT '分类|tech=技术,novel=小说,history=历史,other=其他',
  `description` varchar(255) NOT NULL COMMENT '图书描述',
  `status` varchar(255) NOT NULL COMMENT '状态|draft=草稿,published=已发布,off_shelf=已下架',
  `price` bigint NOT NULL DEFAULT 0 COMMENT '价格(分)',
  `stock_quantity` bigint NOT NULL DEFAULT 0 COMMENT '库存数量',
  `cover_image_url` varchar(255) NOT NULL COMMENT '封面图片URL',
  `tags` varchar(255) NOT NULL COMMENT '标签(逗号分隔)',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
