package service

import (
	"testing"

	"goadmin/core/config"
	corei18n "goadmin/core/i18n"
)

type stubAuthorizationRuntime struct {
	reloadCalls int
	seedCalls   int
	summary     string
}

func (s *stubAuthorizationRuntime) Reload() error {
	s.reloadCalls++
	return nil
}

func (s *stubAuthorizationRuntime) SeedDefaultPolicy() error {
	s.seedCalls++
	return nil
}

func (s *stubAuthorizationRuntime) String() string {
	if s.summary != "" {
		return s.summary
	}
	return "authorization-runtime:stub"
}

func TestServiceUsesAuthorizationRuntimeAbstraction(t *testing.T) {
	t.Parallel()

	stub := &stubAuthorizationRuntime{summary: "authorization-runtime:stub"}
	cfg := &config.Config{}
	cfg.Auth.Casbin.Enabled = true
	cfg.Auth.Casbin.Source = "file"
	cfg.Auth.Casbin.ModelPath = "core/auth/casbin/model/rbac.conf"
	cfg.Auth.Casbin.PolicyPath = "core/auth/casbin/adapter/policy.csv"

	service, err := New(Config{Config: cfg, AuthorizationRuntime: stub})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	status := service.Status()
	if status.Summary != stub.summary {
		t.Fatalf("expected status summary %q, got %q", stub.summary, status.Summary)
	}
	if len(status.LegacyModules) == 0 {
		t.Fatal("expected legacy modules to be reported for migration compatibility")
	}

	if err := service.Reload(); err != nil {
		t.Fatalf("Reload: %v", err)
	}
	if err := service.Seed(); err != nil {
		t.Fatalf("Seed: %v", err)
	}
	if stub.reloadCalls != 1 {
		t.Fatalf("expected 1 reload call, got %d", stub.reloadCalls)
	}
	if stub.seedCalls != 1 {
		t.Fatalf("expected 1 seed call, got %d", stub.seedCalls)
	}
}

func TestServiceWithoutRuntimeStillReturnsGenericStatus(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{}
	cfg.Auth.Casbin.Enabled = false
	cfg.Auth.Casbin.Source = "file"
	cfg.Auth.Casbin.ModelPath = "core/auth/casbin/model/rbac.conf"
	cfg.Auth.Casbin.PolicyPath = "core/auth/casbin/adapter/policy.csv"

	service, err := New(Config{Config: cfg})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	status := service.Status()
	if status.Summary != "authorization module is not configured" {
		t.Fatalf("expected generic fallback summary, got %q", status.Summary)
	}
	if status.Enabled {
		t.Fatal("expected disabled authorization status")
	}
}

func TestServiceUsesLocalizedSummaryWhenAvailable(t *testing.T) {
	t.Parallel()

	if err := corei18n.LoadResourceRoots("../../.."); err != nil {
		t.Fatalf("LoadResourceRoots: %v", err)
	}

	cfg := &config.Config{}
	cfg.Auth.Casbin.Enabled = false
	cfg.Auth.Casbin.Source = "file"
	cfg.Auth.Casbin.ModelPath = "core/auth/casbin/model/rbac.conf"
	cfg.Auth.Casbin.PolicyPath = "core/auth/casbin/adapter/policy.csv"

	service, err := New(Config{Config: cfg})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	status := service.Status()
	if status.Summary != "授权模块未配置" {
		t.Fatalf("expected localized summary, got %q", status.Summary)
	}
}
