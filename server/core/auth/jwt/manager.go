package jwt

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	gjwt "github.com/golang-jwt/jwt/v5"
)

type Config struct {
	Secret          string
	Issuer          string
	Audience        string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

type Manager struct {
	secret     []byte
	issuer     string
	audience   string
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewManager(cfg Config) (*Manager, error) {
	if strings.TrimSpace(cfg.Secret) == "" {
		return nil, fmt.Errorf("jwt secret is required")
	}
	if cfg.AccessTokenTTL <= 0 {
		return nil, fmt.Errorf("jwt access token ttl must be greater than zero")
	}
	if cfg.RefreshTokenTTL <= 0 {
		return nil, fmt.Errorf("jwt refresh token ttl must be greater than zero")
	}
	issuer := strings.TrimSpace(cfg.Issuer)
	if issuer == "" {
		issuer = "GoAdmin"
	}
	audience := strings.TrimSpace(cfg.Audience)
	if audience == "" {
		audience = "goadmin-api"
	}
	return &Manager{
		secret:     []byte(cfg.Secret),
		issuer:     issuer,
		audience:   audience,
		accessTTL:  cfg.AccessTokenTTL,
		refreshTTL: cfg.RefreshTokenTTL,
	}, nil
}

func (m *Manager) IssuePair(identity Identity) (*TokenPair, error) {
	if m == nil {
		return nil, fmt.Errorf("jwt manager is not configured")
	}
	accessToken, accessClaims, err := m.issue(identity, TokenTypeAccess, m.accessTTL)
	if err != nil {
		return nil, err
	}
	refreshToken, refreshClaims, err := m.issue(identity, TokenTypeRefresh, m.refreshTTL)
	if err != nil {
		return nil, err
	}
	return &TokenPair{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		AccessExpiresAt:  accessClaims.ExpiresAt.Time.Unix(),
		RefreshExpiresAt: refreshClaims.ExpiresAt.Time.Unix(),
	}, nil
}

func (m *Manager) ParseAccessToken(tokenString string) (*Claims, error) {
	return m.parse(tokenString, TokenTypeAccess)
}

func (m *Manager) ParseRefreshToken(tokenString string) (*Claims, error) {
	return m.parse(tokenString, TokenTypeRefresh)
}

func (m *Manager) issue(identity Identity, tokenType TokenType, ttl time.Duration) (string, *Claims, error) {
	if m == nil {
		return "", nil, fmt.Errorf("jwt manager is not configured")
	}
	now := time.Now().UTC()
	claims := &Claims{
		TokenType: tokenType,
		Identity:  identity,
		RegisteredClaims: gjwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   identity.UserID,
			Audience:  gjwt.ClaimStrings{m.audience},
			IssuedAt:  gjwt.NewNumericDate(now),
			NotBefore: gjwt.NewNumericDate(now),
			ExpiresAt: gjwt.NewNumericDate(now.Add(ttl)),
			ID:        newTokenID(),
		},
	}
	token := gjwt.NewWithClaims(gjwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", nil, fmt.Errorf("sign jwt: %w", err)
	}
	return signed, claims, nil
}

func (m *Manager) parse(tokenString string, expectedType TokenType) (*Claims, error) {
	if m == nil {
		return nil, fmt.Errorf("jwt manager is not configured")
	}
	claims := &Claims{}
	token, err := gjwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *gjwt.Token) (interface{}, error) {
			if token.Method.Alg() != gjwt.SigningMethodHS256.Alg() {
				return nil, fmt.Errorf("unexpected signing method: %s", token.Header["alg"])
			}
			return m.secret, nil
		},
		gjwt.WithValidMethods([]string{gjwt.SigningMethodHS256.Alg()}),
		gjwt.WithIssuer(m.issuer),
		gjwt.WithAudience(m.audience),
	)
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	if claims.TokenType != expectedType {
		return nil, fmt.Errorf("unexpected token type: %s", claims.TokenType)
	}
	return claims, nil
}

func newTokenID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("%d", time.Now().UTC().UnixNano())
	}
	return hex.EncodeToString(buf)
}
