package service

import (
	"context"
	"fmt"
	"strings"

	corebootstrapcontract "goadmin/core/bootstrap/contract"
	"goadmin/core/config"
	corei18n "goadmin/core/i18n"
)

type Config struct {
	Config               *config.Config
	AuthorizationRuntime corebootstrapcontract.AuthorizationRuntime
}

type Status struct {
	Enabled       bool     `json:"enabled"`
	Source        string   `json:"source"`
	ModelPath     string   `json:"model_path"`
	PolicyPath    string   `json:"policy_path"`
	Summary       string   `json:"summary"`
	LegacyModules []string `json:"legacy_modules,omitempty"`
	Routes        []string `json:"routes,omitempty"`
}

type Service struct {
	cfg                  *config.Config
	authorizationRuntime corebootstrapcontract.AuthorizationRuntime
}

func New(cfg Config) (*Service, error) {
	if cfg.Config == nil {
		return nil, fmt.Errorf("config is required")
	}
	return &Service{cfg: cfg.Config, authorizationRuntime: cfg.AuthorizationRuntime}, nil
}

func (s *Service) Status() Status {
	if s == nil || s.cfg == nil {
		return Status{}
	}
	casbinCfg := s.cfg.Auth.Casbin
	summary := "authorization module is not configured"
	if translated := corei18n.DefaultRegistry().MustTranslate(context.Background(), "casbin.summary.not_configured"); translated != "casbin.summary.not_configured" {
		summary = translated
	}
	if s.authorizationRuntime != nil {
		summary = s.authorizationRuntime.String()
	}
	return Status{
		Enabled:    casbinCfg.Enabled,
		Source:     strings.TrimSpace(casbinCfg.Source),
		ModelPath:  strings.TrimSpace(casbinCfg.ModelPath),
		PolicyPath: strings.TrimSpace(casbinCfg.PolicyPath),
		Summary:    summary,
		LegacyModules: []string{
			"casbin_model",
			"casbin_rule",
		},
		Routes: []string{
			"GET /api/v1/casbin/status",
			"POST /api/v1/casbin/reload",
			"POST /api/v1/casbin/seed",
		},
	}
}

func (s *Service) Reload() error {
	if s == nil || s.authorizationRuntime == nil {
		return nil
	}
	return s.authorizationRuntime.Reload()
}

func (s *Service) Seed() error {
	if s == nil || s.authorizationRuntime == nil {
		return nil
	}
	return s.authorizationRuntime.SeedDefaultPolicy()
}
