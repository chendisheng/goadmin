package model

import "time"

type Book struct {
	Id            string    `json:"id,omitempty" gorm:"column:id;primaryKey;type:varchar(64);size:64;comment:主键ID"`
	TenantId      string    `json:"tenant_id,omitempty" gorm:"column:tenant_id;type:varchar(255);size:255;comment:租户ID"`
	Title         string    `json:"title,omitempty" gorm:"column:title;type:varchar(255);size:255;comment:书名"`
	Author        string    `json:"author,omitempty" gorm:"column:author;type:varchar(255);size:255;comment:作者"`
	Isbn          string    `json:"isbn,omitempty" gorm:"column:isbn;type:varchar(255);size:255;comment:ISBN编号"`
	Publisher     string    `json:"publisher,omitempty" gorm:"column:publisher;type:varchar(255);size:255;comment:出版社"`
	PublishDate   time.Time `json:"publish_date,omitempty" gorm:"column:publish_date;comment:出版日期"`
	Category      string    `json:"category,omitempty" gorm:"column:category;type:varchar(255);size:255;comment:分类|tech=技术,novel=小说,history=历史,other=其他"`
	Description   string    `json:"description,omitempty" gorm:"column:description;type:varchar(255);size:255;comment:图书描述"`
	Status        string    `json:"status,omitempty" gorm:"column:status;type:varchar(255);size:255;comment:状态|draft=草稿,published=已发布,off_shelf=已下架"`
	Price         int64     `json:"price,omitempty" gorm:"column:price;comment:价格(分)"`
	StockQuantity int64     `json:"stock_quantity,omitempty" gorm:"column:stock_quantity;comment:库存数量"`
	CoverImageUrl string    `json:"cover_image_url,omitempty" gorm:"column:cover_image_url;type:varchar(255);size:255;comment:封面图片URL"`
	Tags          string    `json:"tags,omitempty" gorm:"column:tags;type:varchar(255);size:255;comment:标签(逗号分隔)"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (Book) TableName() string {
	return "book"
}

func (m Book) Clone() Book {
	clone := m
	return clone
}
