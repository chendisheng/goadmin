package command

import "time"

type CreateOrder struct {
	TenantId        string    `json:"tenant_id,omitempty"`
	OrderNo         string    `json:"order_no,omitempty"`
	UserId          string    `json:"user_id,omitempty"`
	CustomerName    string    `json:"customer_name,omitempty"`
	CustomerEmail   string    `json:"customer_email,omitempty"`
	CustomerPhone   string    `json:"customer_phone,omitempty"`
	ShippingAddress string    `json:"shipping_address,omitempty"`
	BillingAddress  string    `json:"billing_address,omitempty"`
	OrderStatus     string    `json:"order_status,omitempty"`
	PaymentStatus   string    `json:"payment_status,omitempty"`
	PaymentMethod   string    `json:"payment_method,omitempty"`
	Currency        string    `json:"currency,omitempty"`
	TotalAmount     int64     `json:"total_amount,omitempty"`
	DiscountAmount  int64     `json:"discount_amount,omitempty"`
	TaxAmount       int64     `json:"tax_amount,omitempty"`
	ShippingAmount  int64     `json:"shipping_amount,omitempty"`
	FinalAmount     int64     `json:"final_amount,omitempty"`
	OrderDate       time.Time `json:"order_date,omitempty"`
	ShippedDate     time.Time `json:"shipped_date,omitempty"`
	DeliveredDate   time.Time `json:"delivered_date,omitempty"`
	Notes           string    `json:"notes,omitempty"`
	InternalNotes   string    `json:"internal_notes,omitempty"`
}

type UpdateOrder struct {
	TenantId        string    `json:"tenant_id,omitempty"`
	OrderNo         string    `json:"order_no,omitempty"`
	UserId          string    `json:"user_id,omitempty"`
	CustomerName    string    `json:"customer_name,omitempty"`
	CustomerEmail   string    `json:"customer_email,omitempty"`
	CustomerPhone   string    `json:"customer_phone,omitempty"`
	ShippingAddress string    `json:"shipping_address,omitempty"`
	BillingAddress  string    `json:"billing_address,omitempty"`
	OrderStatus     string    `json:"order_status,omitempty"`
	PaymentStatus   string    `json:"payment_status,omitempty"`
	PaymentMethod   string    `json:"payment_method,omitempty"`
	Currency        string    `json:"currency,omitempty"`
	TotalAmount     int64     `json:"total_amount,omitempty"`
	DiscountAmount  int64     `json:"discount_amount,omitempty"`
	TaxAmount       int64     `json:"tax_amount,omitempty"`
	ShippingAmount  int64     `json:"shipping_amount,omitempty"`
	FinalAmount     int64     `json:"final_amount,omitempty"`
	OrderDate       time.Time `json:"order_date,omitempty"`
	ShippedDate     time.Time `json:"shipped_date,omitempty"`
	DeliveredDate   time.Time `json:"delivered_date,omitempty"`
	Notes           string    `json:"notes,omitempty"`
	InternalNotes   string    `json:"internal_notes,omitempty"`
}
