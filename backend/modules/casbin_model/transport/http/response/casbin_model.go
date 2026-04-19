package response

import "time"

type Item struct {
	Name      string    `json:"name,omitempty"`
	Content   string    `json:"content,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

type List struct {
	Total int64  `json:"total"`
	Items []Item `json:"items"`
}
