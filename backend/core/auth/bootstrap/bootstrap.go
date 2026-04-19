package bootstrap

import (
	"context"
	"fmt"
	"time"

	coreauthcasbinservice "goadmin/core/auth/casbin/service"
	coreauthjwt "goadmin/core/auth/jwt"
	corebootstrapcontract "goadmin/core/bootstrap/contract"
	"goadmin/core/config"

	"gorm.io/gorm"
)

type Authorizer interface {
	EnforceClaims(claims *coreauthjwt.Claims, obj, act string) (bool, error)
}

type RevocationStore interface {
	Revoke(ctx context.Context, jti string, expiresAt time.Time) error
	IsRevoked(ctx context.Context, jti string) (bool, error)
}

type Bundle struct {
	JWT                  *coreauthjwt.Manager
	Authorizer           Authorizer
	AuthorizationRuntime corebootstrapcontract.AuthorizationRuntime
	Casbin               corebootstrapcontract.CasbinRuntime
}

func New(cfg *config.Config, db *gorm.DB) (*Bundle, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}

	accessTTL, refreshTTL, err := cfg.Auth.JWT.Timeouts()
	if err != nil {
		return nil, err
	}

	jwtManager, err := coreauthjwt.NewManager(coreauthjwt.Config{
		Secret:          cfg.Auth.JWT.Secret,
		Issuer:          cfg.Auth.JWT.Issuer,
		Audience:        cfg.Auth.JWT.Audience,
		AccessTokenTTL:  accessTTL,
		RefreshTokenTTL: refreshTTL,
	})
	if err != nil {
		return nil, err
	}

	casbinRuntime, err := coreauthcasbinservice.NewPermissionService(coreauthcasbinservice.Config{
		Enabled:    cfg.Auth.Casbin.Enabled,
		Source:     cfg.Auth.Casbin.Source,
		DB:         db,
		ModelPath:  cfg.Auth.Casbin.ModelPath,
		PolicyPath: cfg.Auth.Casbin.PolicyPath,
	})
	if err != nil {
		return nil, err
	}

	return &Bundle{JWT: jwtManager, Authorizer: casbinRuntime, AuthorizationRuntime: casbinRuntime, Casbin: casbinRuntime}, nil
}
