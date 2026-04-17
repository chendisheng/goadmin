package model

import "time"

type Order struct {
	Id              string    `json:"id,omitempty" gorm:"column:id;primaryKey;type:varchar(64);size:64;comment:主键ID"`
	TenantId        string    `json:"tenant_id,omitempty" gorm:"column:tenant_id;type:varchar(255);size:255;comment:租户ID|tenant"`
	OrderNo         string    `json:"order_no,omitempty" gorm:"column:order_no;type:varchar(255);size:255;comment:订单号"`
	UserId          string    `json:"user_id,omitempty" gorm:"column:user_id;type:varchar(255);size:255;comment:用户ID|ref=user.id"`
	CustomerName    string    `json:"customer_name,omitempty" gorm:"column:customer_name;type:varchar(255);size:255;comment:客户姓名"`
	CustomerEmail   string    `json:"customer_email,omitempty" gorm:"column:customer_email;type:varchar(255);size:255;comment:客户邮箱"`
	CustomerPhone   string    `json:"customer_phone,omitempty" gorm:"column:customer_phone;type:varchar(255);size:255;comment:客户手机号"`
	ShippingAddress string    `json:"shipping_address,omitempty" gorm:"column:shipping_address;type:varchar(255);size:255;comment:收货地址"`
	BillingAddress  string    `json:"billing_address,omitempty" gorm:"column:billing_address;type:varchar(255);size:255;comment:账单地址"`
	OrderStatus     string    `json:"order_status,omitempty" gorm:"column:order_status;type:varchar(255);size:255;comment:订单状态|pending=待处理,paid=已支付,shipped=已发货,delivered=已完成,cancelled=已取消"`
	PaymentStatus   string    `json:"payment_status,omitempty" gorm:"column:payment_status;type:varchar(255);size:255;comment:支付状态|unpaid=未支付,paid=已支付,refunded=已退款,failed=支付失败"`
	PaymentMethod   string    `json:"payment_method,omitempty" gorm:"column:payment_method;type:varchar(255);size:255;comment:支付方式|wechat=微信支付,alipay=支付宝,card=银行卡,paypal=PayPal"`
	Currency        string    `json:"currency,omitempty" gorm:"column:currency;type:varchar(255);size:255;comment:货币类型|CNY=人民币,USD=美元,EUR=欧元,JPY=日元"`
	TotalAmount     int64     `json:"total_amount,omitempty" gorm:"column:total_amount;comment:订单总金额(分)"`
	DiscountAmount  int64     `json:"discount_amount,omitempty" gorm:"column:discount_amount;comment:优惠金额(分)"`
	TaxAmount       int64     `json:"tax_amount,omitempty" gorm:"column:tax_amount;comment:税费(分)"`
	ShippingAmount  int64     `json:"shipping_amount,omitempty" gorm:"column:shipping_amount;comment:运费(分)"`
	FinalAmount     int64     `json:"final_amount,omitempty" gorm:"column:final_amount;comment:实付金额(分)"`
	OrderDate       time.Time `json:"order_date,omitempty" gorm:"column:order_date;comment:下单时间"`
	ShippedDate     time.Time `json:"shipped_date,omitempty" gorm:"column:shipped_date;comment:发货时间"`
	DeliveredDate   time.Time `json:"delivered_date,omitempty" gorm:"column:delivered_date;comment:送达时间"`
	Notes           string    `json:"notes,omitempty" gorm:"column:notes;type:varchar(255);size:255;comment:客户备注"`
	InternalNotes   string    `json:"internal_notes,omitempty" gorm:"column:internal_notes;type:varchar(255);size:255;comment:内部备注"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (m Order) Clone() Order {
	clone := m
	return clone
}
