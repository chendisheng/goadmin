package model

import "time"

type Status string

const (
	StatusActive   Status = "active"
	StatusInactive Status = "inactive"
)

type Role struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id,omitempty"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	Status    Status    `json:"status"`
	Remark    string    `json:"remark,omitempty"`
	MenuIDs   []string  `json:"menu_ids,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (r Role) Clone() Role {
	clone := r
	if r.MenuIDs != nil {
		clone.MenuIDs = append([]string(nil), r.MenuIDs...)
	}
	return clone
}
