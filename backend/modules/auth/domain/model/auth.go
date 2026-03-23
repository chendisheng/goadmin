package model

import "time"

type Credentials struct {
	Username string
	Password string
}

type Identity struct {
	UserID      string
	TenantID    string
	Username    string
	DisplayName string
	Roles       []string
}

type Session struct {
	Identity         Identity
	AccessToken      string
	RefreshToken     string
	AccessExpiresAt  time.Time
	RefreshExpiresAt time.Time
}
