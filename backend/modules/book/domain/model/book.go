package model

import "time"

type Book struct {
	Id            string    `json:"id,omitempty" gorm:"column:id;primaryKey;type:varchar(64);size:64"`
	TenantId      string    `json:"tenant_id,omitempty" gorm:"column:tenant_id;type:varchar(64);size:64"`
	Title         string    `json:"title,omitempty" gorm:"column:title;size:255"`
	Author        string    `json:"author,omitempty" gorm:"column:author;size:255"`
	Isbn          string    `json:"isbn,omitempty" gorm:"column:isbn;size:64"`
	Publisher     string    `json:"publisher,omitempty" gorm:"column:publisher;size:255"`
	PublishDate   time.Time `json:"publish_date,omitempty" gorm:"column:publish_date"`
	Category      string    `json:"category,omitempty" gorm:"column:category;size:128"`
	Description   string    `json:"description,omitempty" gorm:"column:description;type:text"`
	Status        string    `json:"status,omitempty" gorm:"column:status;size:32"`
	Price         int64     `json:"price,omitempty" gorm:"column:price"`
	StockQuantity int64     `json:"stock_quantity,omitempty" gorm:"column:stock_quantity"`
	CoverImageUrl string    `json:"cover_image_url,omitempty" gorm:"column:cover_image_url;size:512"`
	Tags          string    `json:"tags,omitempty" gorm:"column:tags;type:text"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (m Book) Clone() Book {
	clone := m
	return clone
}
