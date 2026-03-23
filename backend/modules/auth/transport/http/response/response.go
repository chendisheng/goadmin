package response

type LoginResponse struct {
	AccessToken      string   `json:"access_token"`
	RefreshToken     string   `json:"refresh_token"`
	TokenType        string   `json:"token_type"`
	ExpiresIn        int64    `json:"expires_in"`
	RefreshExpiresIn int64    `json:"refresh_expires_in"`
	User             UserInfo `json:"user"`
}

type UserInfo struct {
	UserID      string   `json:"user_id"`
	TenantID    string   `json:"tenant_id,omitempty"`
	Username    string   `json:"username"`
	DisplayName string   `json:"display_name,omitempty"`
	Roles       []string `json:"roles,omitempty"`
}
