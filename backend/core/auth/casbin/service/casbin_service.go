package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	casbinadapter "goadmin/core/auth/casbin/adapter"
	coreauthjwt "goadmin/core/auth/jwt"
	coretenant "goadmin/core/tenant"
)

func (s *PermissionService) EnforceClaims(claims *coreauthjwt.Claims, obj, act string) (bool, error) {
	if claims == nil {
		return false, nil
	}
	if !coretenant.Enabled() {
		return s.EnforceRoles(claims.Roles, obj, act)
	}
	return s.enforceSubjects(claims.TenantID, claims.Username, claims.Roles, obj, act)
}

func (s *PermissionService) EnforceRoles(roles []string, obj, act string) (bool, error) {
	return s.enforceSubjects("", "", roles, obj, act)
}

func (s *PermissionService) enforceSubjects(tenantID, username string, roles []string, obj, act string) (bool, error) {
	if s == nil || s.enforcer == nil {
		return true, nil
	}
	if len(roles) == 0 {
		return false, nil
	}
	tenantID = normalize(tenantID)
	username = normalize(username)
	for _, role := range roles {
		role = normalize(role)
		if role == "" {
			continue
		}
		for _, subject := range candidateSubjects(tenantID, username, role) {
			ok, err := s.enforcer.Enforce(subject, obj, act)
			if err != nil {
				return false, err
			}
			if ok {
				return true, nil
			}
		}
	}
	return false, nil
}

func (s *PermissionService) Reload() error {
	if s == nil || s.enforcer == nil {
		return nil
	}
	return s.enforcer.Reload()
}

func (s *PermissionService) GrantPolicy(rule casbinadapter.Rule) error {
	if s == nil || s.enforcer == nil {
		return nil
	}
	return s.enforcer.AddPolicy(rule)
}

func (s *PermissionService) RevokePolicy(rule casbinadapter.Rule) error {
	if s == nil || s.enforcer == nil {
		return nil
	}
	return s.enforcer.RemovePolicy(rule)
}

func (s *PermissionService) SeedDefaultPolicy() error {
	if s == nil || s.enforcer == nil {
		return nil
	}
	for _, rule := range defaultRules() {
		if err := s.enforcer.AddPolicy(rule); err != nil {
			return err
		}
	}
	return nil
}

func (s *PermissionService) WithReloadOnSignal(ctx context.Context, cb func() error) error {
	if s == nil || cb == nil {
		return nil
	}
	go func() {
		<-ctx.Done()
		_ = cb()
	}()
	return nil
}

func normalize(value string) string {
	return strings.TrimSpace(value)
}

func candidateSubjects(tenantID, username, role string) []string {
	seen := make(map[string]struct{}, 3)
	add := func(value string, out *[]string) {
		value = strings.TrimSpace(value)
		if value == "" {
			return
		}
		if _, ok := seen[value]; ok {
			return
		}
		seen[value] = struct{}{}
		*out = append(*out, value)
	}

	result := make([]string, 0, 3)
	if tenantID != "" && username != "" {
		add(fmt.Sprintf("%s:%s:%s", tenantID, username, role), &result)
	}
	if tenantID != "" {
		add(fmt.Sprintf("%s:%s", tenantID, role), &result)
	}
	add(role, &result)
	return result
}

func defaultRules() []casbinadapter.Rule {
	return []casbinadapter.Rule{
		{Subject: "admin", Object: "/api/v1/auth/me", Action: "GET"},
		{Subject: "admin", Object: "/api/v1/auth/logout", Action: "POST"},
		{Subject: "user", Object: "/api/v1/auth/me", Action: "GET"},
		{Subject: "user", Object: "/api/v1/auth/logout", Action: "POST"},
	}
}

var _ = time.Second
