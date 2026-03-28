package jwt

import (
	"errors"
	"fmt"
	"strings"

	gjwt "github.com/golang-jwt/jwt/v5"
)

type TokenType string

const (
	TokenTypeAccess  TokenType = "access"
	TokenTypeRefresh TokenType = "refresh"
)

type Identity struct {
	UserID      string   `json:"user_id"`
	TenantID    string   `json:"tenant_id,omitempty"`
	Username    string   `json:"username"`
	DisplayName string   `json:"display_name,omitempty"`
	Roles       []string `json:"roles,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
}

type Claims struct {
	TokenType TokenType `json:"token_type"`
	Identity
	gjwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken      string
	RefreshToken     string
	AccessExpiresAt  int64
	RefreshExpiresAt int64
}

func (c Claims) Validate() error {
	if strings.TrimSpace(c.UserID) == "" {
		return errors.New("user_id is required")
	}
	if strings.TrimSpace(c.Username) == "" {
		return errors.New("username is required")
	}
	if strings.TrimSpace(string(c.TokenType)) == "" {
		return errors.New("token_type is required")
	}
	if c.TokenType != TokenTypeAccess && c.TokenType != TokenTypeRefresh {
		return fmt.Errorf("invalid token_type: %s", c.TokenType)
	}
	return nil
}

func (c Claims) PrimarySubject() string {
	if strings.TrimSpace(c.Username) != "" {
		return c.Username
	}
	return c.UserID
}
