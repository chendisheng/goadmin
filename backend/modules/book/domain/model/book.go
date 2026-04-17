package model

import "time"

type Book struct {
	Id            string    `json:"id,omitempty" gorm:"column:id;primaryKey;type:varchar(64);size:64"`
	TenantId      string    `json:"tenant_id,omitempty" gorm:"column:tenant_id;type:varchar(255);size:255"`
	Title         string    `json:"title,omitempty" gorm:"column:title;type:varchar(255);size:255"`
	Author        string    `json:"author,omitempty" gorm:"column:author;type:varchar(255);size:255"`
	Isbn          string    `json:"isbn,omitempty" gorm:"column:isbn;type:varchar(255);size:255"`
	Publisher     string    `json:"publisher,omitempty" gorm:"column:publisher;type:varchar(255);size:255"`
	PublishDate   time.Time `json:"publish_date,omitempty" gorm:"column:publish_date"`
	Category      string    `json:"category,omitempty" gorm:"column:category;type:varchar(255);size:255"`
	Description   string    `json:"description,omitempty" gorm:"column:description;type:varchar(255);size:255"`
	Status        string    `json:"status,omitempty" gorm:"column:status;type:varchar(255);size:255"`
	Price         int64     `json:"price,omitempty" gorm:"column:price"`
	StockQuantity int64     `json:"stock_quantity,omitempty" gorm:"column:stock_quantity"`
	CoverImageUrl string    `json:"cover_image_url,omitempty" gorm:"column:cover_image_url;type:varchar(255);size:255"`
	Tags          string    `json:"tags,omitempty" gorm:"column:tags;type:varchar(255);size:255"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (m Book) Clone() Book {
	clone := m
	return clone
}
