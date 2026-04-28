package response

import "time"

type CategoryItem struct {
	ID          string    `json:"id,omitempty"`
	Code        string    `json:"code,omitempty"`
	Name        string    `json:"name,omitempty"`
	Description string    `json:"description,omitempty"`
	Status      string    `json:"status,omitempty"`
	Sort        int       `json:"sort,omitempty"`
	Remark      string    `json:"remark,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CategoryList struct {
	Total int64          `json:"total"`
	Items []CategoryItem `json:"items"`
}

type Item struct {
	ID         string    `json:"id,omitempty"`
	CategoryID string    `json:"category_id,omitempty"`
	Value      string    `json:"value,omitempty"`
	Label      string    `json:"label,omitempty"`
	TagType    string    `json:"tag_type,omitempty"`
	TagColor   string    `json:"tag_color,omitempty"`
	Extra      string    `json:"extra,omitempty"`
	IsDefault  bool      `json:"is_default,omitempty"`
	Status     string    `json:"status,omitempty"`
	Sort       int       `json:"sort,omitempty"`
	Remark     string    `json:"remark,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type ItemList struct {
	Total int64  `json:"total"`
	Items []Item `json:"items"`
}

type Lookup struct {
	Items []Item `json:"items"`
}
