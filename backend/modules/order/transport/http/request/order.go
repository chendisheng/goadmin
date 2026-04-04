package request

import "time"

type ListRequest struct {
	Keyword  string `json:"keyword,omitempty" form:"keyword"`
	Page     int    `json:"page,omitempty" form:"page"`
	PageSize int    `json:"page_size,omitempty" form:"page_size"`
}

type CreateRequest struct {
	TenantId        string    `json:"tenant_id,omitempty" form:"tenant_id"`
	OrderNo         string    `json:"order_no,omitempty" form:"order_no"`
	UserId          string    `json:"user_id,omitempty" form:"user_id"`
	CustomerName    string    `json:"customer_name,omitempty" form:"customer_name"`
	CustomerEmail   string    `json:"customer_email,omitempty" form:"customer_email"`
	CustomerPhone   string    `json:"customer_phone,omitempty" form:"customer_phone"`
	ShippingAddress string    `json:"shipping_address,omitempty" form:"shipping_address"`
	BillingAddress  string    `json:"billing_address,omitempty" form:"billing_address"`
	OrderStatus     string    `json:"order_status,omitempty" form:"order_status"`
	PaymentStatus   string    `json:"payment_status,omitempty" form:"payment_status"`
	PaymentMethod   string    `json:"payment_method,omitempty" form:"payment_method"`
	Currency        string    `json:"currency,omitempty" form:"currency"`
	TotalAmount     int64     `json:"total_amount,omitempty" form:"total_amount"`
	DiscountAmount  int64     `json:"discount_amount,omitempty" form:"discount_amount"`
	TaxAmount       int64     `json:"tax_amount,omitempty" form:"tax_amount"`
	ShippingAmount  int64     `json:"shipping_amount,omitempty" form:"shipping_amount"`
	FinalAmount     int64     `json:"final_amount,omitempty" form:"final_amount"`
	OrderDate       time.Time `json:"order_date,omitempty" form:"order_date"`
	ShippedDate     time.Time `json:"shipped_date,omitempty" form:"shipped_date"`
	DeliveredDate   time.Time `json:"delivered_date,omitempty" form:"delivered_date"`
	Notes           string    `json:"notes,omitempty" form:"notes"`
	InternalNotes   string    `json:"internal_notes,omitempty" form:"internal_notes"`
}

type UpdateRequest struct {
	TenantId        string    `json:"tenant_id,omitempty" form:"tenant_id"`
	OrderNo         string    `json:"order_no,omitempty" form:"order_no"`
	UserId          string    `json:"user_id,omitempty" form:"user_id"`
	CustomerName    string    `json:"customer_name,omitempty" form:"customer_name"`
	CustomerEmail   string    `json:"customer_email,omitempty" form:"customer_email"`
	CustomerPhone   string    `json:"customer_phone,omitempty" form:"customer_phone"`
	ShippingAddress string    `json:"shipping_address,omitempty" form:"shipping_address"`
	BillingAddress  string    `json:"billing_address,omitempty" form:"billing_address"`
	OrderStatus     string    `json:"order_status,omitempty" form:"order_status"`
	PaymentStatus   string    `json:"payment_status,omitempty" form:"payment_status"`
	PaymentMethod   string    `json:"payment_method,omitempty" form:"payment_method"`
	Currency        string    `json:"currency,omitempty" form:"currency"`
	TotalAmount     int64     `json:"total_amount,omitempty" form:"total_amount"`
	DiscountAmount  int64     `json:"discount_amount,omitempty" form:"discount_amount"`
	TaxAmount       int64     `json:"tax_amount,omitempty" form:"tax_amount"`
	ShippingAmount  int64     `json:"shipping_amount,omitempty" form:"shipping_amount"`
	FinalAmount     int64     `json:"final_amount,omitempty" form:"final_amount"`
	OrderDate       time.Time `json:"order_date,omitempty" form:"order_date"`
	ShippedDate     time.Time `json:"shipped_date,omitempty" form:"shipped_date"`
	DeliveredDate   time.Time `json:"delivered_date,omitempty" form:"delivered_date"`
	Notes           string    `json:"notes,omitempty" form:"notes"`
	InternalNotes   string    `json:"internal_notes,omitempty" form:"internal_notes"`
}
