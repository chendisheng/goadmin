package model

import "time"

type Status string

const (
	StatusActive   Status = "active"
	StatusInactive Status = "inactive"
)

type User struct {
	ID           string    `json:"id"`
	TenantID     string    `json:"tenant_id,omitempty"`
	Username     string    `json:"username"`
	DisplayName  string    `json:"display_name,omitempty"`
	Language     string    `json:"language,omitempty"`
	Mobile       string    `json:"mobile,omitempty"`
	Email        string    `json:"email,omitempty"`
	Status       Status    `json:"status"`
	RoleCodes    []string  `json:"role_codes,omitempty"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (User) TableName() string {
	return "user"
}

func (u User) Clone() User {
	clone := u
	if u.RoleCodes != nil {
		clone.RoleCodes = append([]string(nil), u.RoleCodes...)
	}
	return clone
}
