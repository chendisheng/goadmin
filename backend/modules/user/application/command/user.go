package command

type CreateUser struct {
	TenantID     string   `json:"tenant_id,omitempty"`
	Username     string   `json:"username"`
	DisplayName  string   `json:"display_name,omitempty"`
	Mobile       string   `json:"mobile,omitempty"`
	Email        string   `json:"email,omitempty"`
	Status       string   `json:"status,omitempty"`
	RoleCodes    []string `json:"role_codes,omitempty"`
	PasswordHash string   `json:"password_hash,omitempty"`
}

type UpdateUser struct {
	TenantID     string   `json:"tenant_id,omitempty"`
	Username     string   `json:"username,omitempty"`
	DisplayName  string   `json:"display_name,omitempty"`
	Mobile       string   `json:"mobile,omitempty"`
	Email        string   `json:"email,omitempty"`
	Status       string   `json:"status,omitempty"`
	RoleCodes    []string `json:"role_codes,omitempty"`
	PasswordHash string   `json:"password_hash,omitempty"`
}
