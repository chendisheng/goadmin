package model

import "time"

type Book struct {
	Id            string    `json:"id,omitempty" gorm:"column:id;primaryKey"`
	TenantId      string    `json:"tenant_id,omitempty" gorm:"column:tenant_id"`
	Title         string    `json:"title,omitempty" gorm:"column:title"`
	Author        string    `json:"author,omitempty" gorm:"column:author"`
	Isbn          string    `json:"isbn,omitempty" gorm:"column:isbn"`
	Publisher     string    `json:"publisher,omitempty" gorm:"column:publisher"`
	PublishDate   time.Time `json:"publish_date,omitempty" gorm:"column:publish_date"`
	Category      string    `json:"category,omitempty" gorm:"column:category"`
	Description   string    `json:"description,omitempty" gorm:"column:description"`
	Status        string    `json:"status,omitempty" gorm:"column:status"`
	Price         int64     `json:"price,omitempty" gorm:"column:price"`
	StockQuantity int64     `json:"stock_quantity,omitempty" gorm:"column:stock_quantity"`
	CoverImageUrl string    `json:"cover_image_url,omitempty" gorm:"column:cover_image_url"`
	Tags          string    `json:"tags,omitempty" gorm:"column:tags"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (m Book) Clone() Book {
	clone := m
	return clone
}
