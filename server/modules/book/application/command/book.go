package command

import "time"

type CreateBook struct {
	TenantId      string    `json:"tenant_id,omitempty"`
	Title         string    `json:"title,omitempty"`
	Author        string    `json:"author,omitempty"`
	Isbn          string    `json:"isbn,omitempty"`
	Publisher     string    `json:"publisher,omitempty"`
	PublishDate   time.Time `json:"publish_date,omitempty"`
	Category      string    `json:"category,omitempty"`
	Description   string    `json:"description,omitempty"`
	Status        string    `json:"status,omitempty"`
	Price         int64     `json:"price,omitempty"`
	StockQuantity int64     `json:"stock_quantity,omitempty"`
	CoverImageUrl string    `json:"cover_image_url,omitempty"`
	Tags          string    `json:"tags,omitempty"`
}

type UpdateBook struct {
	TenantId      string    `json:"tenant_id,omitempty"`
	Title         string    `json:"title,omitempty"`
	Author        string    `json:"author,omitempty"`
	Isbn          string    `json:"isbn,omitempty"`
	Publisher     string    `json:"publisher,omitempty"`
	PublishDate   time.Time `json:"publish_date,omitempty"`
	Category      string    `json:"category,omitempty"`
	Description   string    `json:"description,omitempty"`
	Status        string    `json:"status,omitempty"`
	Price         int64     `json:"price,omitempty"`
	StockQuantity int64     `json:"stock_quantity,omitempty"`
	CoverImageUrl string    `json:"cover_image_url,omitempty"`
	Tags          string    `json:"tags,omitempty"`
}
