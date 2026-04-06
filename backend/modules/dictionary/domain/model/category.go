package model

import "time"

type Category struct {
	ID          string    `json:"id" gorm:"column:id;primaryKey;type:varchar(64);size:64"`
	Code        string    `json:"code" gorm:"column:code;size:64;not null;uniqueIndex"`
	Name        string    `json:"name" gorm:"column:name;size:128;not null;index"`
	Description string    `json:"description,omitempty" gorm:"column:description;type:text"`
	Status      Status    `json:"status" gorm:"column:status;size:32;not null;default:enabled;index"`
	Sort        int       `json:"sort" gorm:"column:sort;index"`
	Remark      string    `json:"remark,omitempty" gorm:"column:remark;size:512"`
	CreatedAt   time.Time `json:"created_at" gorm:"column:created_at"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"column:updated_at"`
}

func (c Category) Clone() Category {
	return c
}
