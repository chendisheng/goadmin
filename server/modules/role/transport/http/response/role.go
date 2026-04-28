package response

import "time"

type Item struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id,omitempty"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	Status    string    `json:"status"`
	Remark    string    `json:"remark,omitempty"`
	MenuIDs   []string  `json:"menu_ids,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type List struct {
	Total int64  `json:"total"`
	Items []Item `json:"items"`
}
