package service

import (
	"context"
	"fmt"
	"time"

	coreauthbootstrap "goadmin/core/auth/bootstrap"
	coreauthjwt "goadmin/core/auth/jwt"
	apperrors "goadmin/core/errors"
	coretenant "goadmin/core/tenant"
	"goadmin/modules/auth/domain/model"
)

type CredentialStore interface {
	Authenticate(ctx context.Context, username, password string) (model.Identity, error)
}

type Service struct {
	tokens      *coreauthjwt.Manager
	authorizer  coreauthbootstrap.Authorizer
	credentials CredentialStore
	revocations coreauthbootstrap.RevocationStore
}

func New(tokens *coreauthjwt.Manager, authorizer coreauthbootstrap.Authorizer, credentials CredentialStore, revocations coreauthbootstrap.RevocationStore) (*Service, error) {
	if tokens == nil {
		return nil, fmt.Errorf("jwt manager is required")
	}
	if authorizer == nil {
		return nil, fmt.Errorf("authorizer is required")
	}
	if credentials == nil {
		return nil, fmt.Errorf("credential store is required")
	}
	if revocations == nil {
		revocations = NewMemoryRevocationStore()
	}
	return &Service{
		tokens:      tokens,
		authorizer:  authorizer,
		credentials: credentials,
		revocations: revocations,
	}, nil
}

func (s *Service) Login(ctx context.Context, input model.Credentials) (*model.Session, error) {
	identity, err := s.credentials.Authenticate(ctx, input.Username, input.Password)
	if err != nil {
		return nil, err
	}
	if !coretenant.Enabled() {
		identity.TenantID = ""
	}

	pair, err := s.tokens.IssuePair(coreauthjwt.Identity{
		UserID:      identity.UserID,
		TenantID:    identity.TenantID,
		Username:    identity.Username,
		DisplayName: identity.DisplayName,
		Language:    identity.Language,
		Roles:       append([]string(nil), identity.Roles...),
		Permissions: append([]string(nil), identity.Permissions...),
	})
	if err != nil {
		return nil, apperrors.Wrap(err, apperrors.CodeInternal, "issue token failed")
	}

	return &model.Session{
		Identity:         identity,
		AccessToken:      pair.AccessToken,
		RefreshToken:     pair.RefreshToken,
		AccessExpiresAt:  time.Unix(pair.AccessExpiresAt, 0).UTC(),
		RefreshExpiresAt: time.Unix(pair.RefreshExpiresAt, 0).UTC(),
	}, nil
}

func (s *Service) Me(_ context.Context, claims *coreauthjwt.Claims) (*model.Identity, error) {
	if claims == nil {
		return nil, apperrors.New(apperrors.CodeUnauthorized, "authentication required")
	}
	tenantID := claims.TenantID
	if !coretenant.Enabled() {
		tenantID = ""
	}
	identity := &model.Identity{
		UserID:      claims.UserID,
		TenantID:    tenantID,
		Username:    claims.Username,
		DisplayName: claims.DisplayName,
		Language:    claims.Language,
		Roles:       append([]string(nil), claims.Roles...),
		Permissions: append([]string(nil), claims.Permissions...),
	}
	return identity, nil
}

func (s *Service) Logout(ctx context.Context, claims *coreauthjwt.Claims) error {
	if claims == nil {
		return apperrors.New(apperrors.CodeUnauthorized, "authentication required")
	}
	if s.revocations == nil {
		return nil
	}
	return s.revocations.Revoke(ctx, claims.RegisteredClaims.ID, claims.RegisteredClaims.ExpiresAt.Time)
}

func (s *Service) Authorize(claims *coreauthjwt.Claims, obj, act string) (bool, error) {
	if claims == nil {
		return false, apperrors.New(apperrors.CodeUnauthorized, "authentication required")
	}
	return s.authorizer.EnforceClaims(claims, obj, act)
}
