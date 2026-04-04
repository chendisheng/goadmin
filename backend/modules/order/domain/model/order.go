package model

import "time"

type Order struct {
	Id              string    `json:"id,omitempty" gorm:"column:id;primaryKey"`
	TenantId        string    `json:"tenant_id,omitempty" gorm:"column:tenant_id"`
	OrderNo         string    `json:"order_no,omitempty" gorm:"column:order_no"`
	UserId          string    `json:"user_id,omitempty" gorm:"column:user_id"`
	CustomerName    string    `json:"customer_name,omitempty" gorm:"column:customer_name"`
	CustomerEmail   string    `json:"customer_email,omitempty" gorm:"column:customer_email"`
	CustomerPhone   string    `json:"customer_phone,omitempty" gorm:"column:customer_phone"`
	ShippingAddress string    `json:"shipping_address,omitempty" gorm:"column:shipping_address"`
	BillingAddress  string    `json:"billing_address,omitempty" gorm:"column:billing_address"`
	OrderStatus     string    `json:"order_status,omitempty" gorm:"column:order_status"`
	PaymentStatus   string    `json:"payment_status,omitempty" gorm:"column:payment_status"`
	PaymentMethod   string    `json:"payment_method,omitempty" gorm:"column:payment_method"`
	Currency        string    `json:"currency,omitempty" gorm:"column:currency"`
	TotalAmount     int64     `json:"total_amount,omitempty" gorm:"column:total_amount"`
	DiscountAmount  int64     `json:"discount_amount,omitempty" gorm:"column:discount_amount"`
	TaxAmount       int64     `json:"tax_amount,omitempty" gorm:"column:tax_amount"`
	ShippingAmount  int64     `json:"shipping_amount,omitempty" gorm:"column:shipping_amount"`
	FinalAmount     int64     `json:"final_amount,omitempty" gorm:"column:final_amount"`
	OrderDate       time.Time `json:"order_date,omitempty" gorm:"column:order_date"`
	ShippedDate     time.Time `json:"shipped_date,omitempty" gorm:"column:shipped_date"`
	DeliveredDate   time.Time `json:"delivered_date,omitempty" gorm:"column:delivered_date"`
	Notes           string    `json:"notes,omitempty" gorm:"column:notes"`
	InternalNotes   string    `json:"internal_notes,omitempty" gorm:"column:internal_notes"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (m Order) Clone() Order {
	clone := m
	return clone
}
