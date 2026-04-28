package response

import "time"

type Item struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id,omitempty"`
	Username    string    `json:"username"`
	DisplayName string    `json:"display_name,omitempty"`
	Language    string    `json:"language,omitempty"`
	Mobile      string    `json:"mobile,omitempty"`
	Email       string    `json:"email,omitempty"`
	Status      string    `json:"status"`
	RoleCodes   []string  `json:"role_codes,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type List struct {
	Total int64  `json:"total"`
	Items []Item `json:"items"`
}
