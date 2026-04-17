-- Auto-generated schema for orders
-- Database: goadmin
CREATE TABLE IF NOT EXISTS `orders` (
  `id` varchar(64) NOT NULL COMMENT '主键ID',
  `tenant_id` varchar(255) NOT NULL COMMENT '租户ID|tenant',
  `order_no` varchar(255) NOT NULL COMMENT '订单号',
  `user_id` varchar(255) NOT NULL COMMENT '用户ID|ref=user.id',
  `customer_name` varchar(255) NOT NULL COMMENT '客户姓名',
  `customer_email` varchar(255) NOT NULL COMMENT '客户邮箱',
  `customer_phone` varchar(255) NOT NULL COMMENT '客户手机号',
  `shipping_address` varchar(255) NOT NULL COMMENT '收货地址',
  `billing_address` varchar(255) NOT NULL COMMENT '账单地址',
  `order_status` varchar(255) NOT NULL COMMENT '订单状态|pending=待处理,paid=已支付,shipped=已发货,delivered=已完成,cancelled=已取消',
  `payment_status` varchar(255) NOT NULL COMMENT '支付状态|unpaid=未支付,paid=已支付,refunded=已退款,failed=支付失败',
  `payment_method` varchar(255) NOT NULL COMMENT '支付方式|wechat=微信支付,alipay=支付宝,card=银行卡,paypal=PayPal',
  `currency` varchar(255) NOT NULL COMMENT '货币类型|CNY=人民币,USD=美元,EUR=欧元,JPY=日元',
  `total_amount` bigint NOT NULL DEFAULT 0 COMMENT '订单总金额(分)',
  `discount_amount` bigint NOT NULL DEFAULT 0 COMMENT '优惠金额(分)',
  `tax_amount` bigint NOT NULL DEFAULT 0 COMMENT '税费(分)',
  `shipping_amount` bigint NOT NULL DEFAULT 0 COMMENT '运费(分)',
  `final_amount` bigint NOT NULL DEFAULT 0 COMMENT '实付金额(分)',
  `order_date` datetime NULL COMMENT '下单时间',
  `shipped_date` datetime NULL COMMENT '发货时间',
  `delivered_date` datetime NULL COMMENT '送达时间',
  `notes` varchar(255) NOT NULL COMMENT '客户备注',
  `internal_notes` varchar(255) NOT NULL COMMENT '内部备注',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
