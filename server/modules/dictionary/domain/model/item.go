package model

import "time"

type Item struct {
	ID         string    `json:"id" gorm:"column:id;primaryKey;type:varchar(64);size:64"`
	CategoryID string    `json:"category_id" gorm:"column:category_id;type:varchar(64);size:64;index"`
	Value      string    `json:"value" gorm:"column:value;size:128;not null;index"`
	Label      string    `json:"label" gorm:"column:label;size:128;not null;index"`
	TagType    string    `json:"tag_type,omitempty" gorm:"column:tag_type;size:32"`
	TagColor   string    `json:"tag_color,omitempty" gorm:"column:tag_color;size:32"`
	Extra      string    `json:"extra,omitempty" gorm:"column:extra;type:text"`
	IsDefault  bool      `json:"is_default" gorm:"column:is_default;default:false;index"`
	Status     Status    `json:"status" gorm:"column:status;size:32;not null;default:enabled;index"`
	Sort       int       `json:"sort" gorm:"column:sort;index"`
	Remark     string    `json:"remark,omitempty" gorm:"column:remark;size:512"`
	CreatedAt  time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"column:updated_at"`
}

func (Item) TableName() string {
	return "dictionary_item"
}

func (i Item) Clone() Item {
	return i
}
