package model

import "time"

type Order struct {
	Id              string    `json:"id,omitempty" gorm:"column:id;primaryKey;type:varchar(64);size:64"`
	TenantId        string    `json:"tenant_id,omitempty" gorm:"column:tenant_id;type:varchar(255);size:255"`
	OrderNo         string    `json:"order_no,omitempty" gorm:"column:order_no;type:varchar(255);size:255"`
	UserId          string    `json:"user_id,omitempty" gorm:"column:user_id;type:varchar(255);size:255"`
	CustomerName    string    `json:"customer_name,omitempty" gorm:"column:customer_name;type:varchar(255);size:255"`
	CustomerEmail   string    `json:"customer_email,omitempty" gorm:"column:customer_email;type:varchar(255);size:255"`
	CustomerPhone   string    `json:"customer_phone,omitempty" gorm:"column:customer_phone;type:varchar(255);size:255"`
	ShippingAddress string    `json:"shipping_address,omitempty" gorm:"column:shipping_address;type:varchar(255);size:255"`
	BillingAddress  string    `json:"billing_address,omitempty" gorm:"column:billing_address;type:varchar(255);size:255"`
	OrderStatus     string    `json:"order_status,omitempty" gorm:"column:order_status;type:varchar(255);size:255"`
	PaymentStatus   string    `json:"payment_status,omitempty" gorm:"column:payment_status;type:varchar(255);size:255"`
	PaymentMethod   string    `json:"payment_method,omitempty" gorm:"column:payment_method;type:varchar(255);size:255"`
	Currency        string    `json:"currency,omitempty" gorm:"column:currency;type:varchar(255);size:255"`
	TotalAmount     int64     `json:"total_amount,omitempty" gorm:"column:total_amount"`
	DiscountAmount  int64     `json:"discount_amount,omitempty" gorm:"column:discount_amount"`
	TaxAmount       int64     `json:"tax_amount,omitempty" gorm:"column:tax_amount"`
	ShippingAmount  int64     `json:"shipping_amount,omitempty" gorm:"column:shipping_amount"`
	FinalAmount     int64     `json:"final_amount,omitempty" gorm:"column:final_amount"`
	OrderDate       time.Time `json:"order_date,omitempty" gorm:"column:order_date"`
	ShippedDate     time.Time `json:"shipped_date,omitempty" gorm:"column:shipped_date"`
	DeliveredDate   time.Time `json:"delivered_date,omitempty" gorm:"column:delivered_date"`
	Notes           string    `json:"notes,omitempty" gorm:"column:notes;type:varchar(255);size:255"`
	InternalNotes   string    `json:"internal_notes,omitempty" gorm:"column:internal_notes;type:varchar(255);size:255"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (m Order) Clone() Order {
	clone := m
	return clone
}
