package request

import "time"

type ListRequest struct {
	Keyword  string `json:"keyword,omitempty" form:"keyword"`
	Page     int    `json:"page,omitempty" form:"page"`
	PageSize int    `json:"page_size,omitempty" form:"page_size"`
}

type CreateRequest struct {
	TenantId      string    `json:"tenant_id,omitempty" form:"tenant_id"`
	Title         string    `json:"title,omitempty" form:"title"`
	Author        string    `json:"author,omitempty" form:"author"`
	Isbn          string    `json:"isbn,omitempty" form:"isbn"`
	Publisher     string    `json:"publisher,omitempty" form:"publisher"`
	PublishDate   time.Time `json:"publish_date,omitempty" form:"publish_date"`
	Category      string    `json:"category,omitempty" form:"category"`
	Description   string    `json:"description,omitempty" form:"description"`
	Status        string    `json:"status,omitempty" form:"status"`
	Price         int64     `json:"price,omitempty" form:"price"`
	StockQuantity int64     `json:"stock_quantity,omitempty" form:"stock_quantity"`
	CoverImageUrl string    `json:"cover_image_url,omitempty" form:"cover_image_url"`
	Tags          string    `json:"tags,omitempty" form:"tags"`
}

type UpdateRequest struct {
	TenantId      string    `json:"tenant_id,omitempty" form:"tenant_id"`
	Title         string    `json:"title,omitempty" form:"title"`
	Author        string    `json:"author,omitempty" form:"author"`
	Isbn          string    `json:"isbn,omitempty" form:"isbn"`
	Publisher     string    `json:"publisher,omitempty" form:"publisher"`
	PublishDate   time.Time `json:"publish_date,omitempty" form:"publish_date"`
	Category      string    `json:"category,omitempty" form:"category"`
	Description   string    `json:"description,omitempty" form:"description"`
	Status        string    `json:"status,omitempty" form:"status"`
	Price         int64     `json:"price,omitempty" form:"price"`
	StockQuantity int64     `json:"stock_quantity,omitempty" form:"stock_quantity"`
	CoverImageUrl string    `json:"cover_image_url,omitempty" form:"cover_image_url"`
	Tags          string    `json:"tags,omitempty" form:"tags"`
}
