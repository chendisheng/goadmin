package request

type ListRequest struct {
	TenantID string `form:"tenant_id"`
	Keyword  string `form:"keyword"`
	Status   string `form:"status"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}

type CreateRequest struct {
	TenantID     string   `json:"tenant_id"`
	Username     string   `json:"username" binding:"required"`
	DisplayName  string   `json:"display_name"`
	Mobile       string   `json:"mobile"`
	Email        string   `json:"email"`
	Status       string   `json:"status"`
	RoleCodes    []string `json:"role_codes"`
	PasswordHash string   `json:"password_hash"`
}

type UpdateRequest struct {
	TenantID     string   `json:"tenant_id"`
	Username     string   `json:"username"`
	DisplayName  string   `json:"display_name"`
	Mobile       string   `json:"mobile"`
	Email        string   `json:"email"`
	Status       string   `json:"status"`
	RoleCodes    []string `json:"role_codes"`
	PasswordHash string   `json:"password_hash"`
}
