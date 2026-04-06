package service

import (
	"os"
	"path/filepath"
	"testing"

	coreauthjwt "goadmin/core/auth/jwt"
	coretenant "goadmin/core/tenant"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestPermissionServiceEnforceClaimsWithTenantSubject(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	policyPath := filepath.Join(tmpDir, "policy.csv")
	policy := "p, system:*:admin, /api/v1/users, GET\n"
	if err := os.WriteFile(policyPath, []byte(policy), 0o600); err != nil {
		t.Fatalf("write policy file: %v", err)
	}

	service, err := NewPermissionService(Config{
		Enabled:    true,
		ModelPath:  filepath.Clean(filepath.Join("..", "model", "rbac.conf")),
		PolicyPath: policyPath,
	})
	if err != nil {
		t.Fatalf("NewPermissionService: %v", err)
	}

	allowed, err := service.EnforceClaims(&coreauthjwt.Claims{
		Identity: coreauthjwt.Identity{
			TenantID: "system",
			Username: "alice",
			Roles:    []string{"admin"},
		},
	}, "/api/v1/users", "GET")
	if err != nil {
		t.Fatalf("EnforceClaims returned error: %v", err)
	}
	if !allowed {
		t.Fatal("expected tenant-aware policy to allow access")
	}

	allowed, err = service.EnforceClaims(&coreauthjwt.Claims{
		Identity: coreauthjwt.Identity{
			TenantID: "tenant-b",
			Username: "alice",
			Roles:    []string{"admin"},
		},
	}, "/api/v1/users", "GET")
	if err != nil {
		t.Fatalf("EnforceClaims returned error: %v", err)
	}
	if allowed {
		t.Fatal("expected non-matching tenant to be denied")
	}
}

func TestPermissionServiceEnforceClaimsWhenTenantDisabledUsesRoleMode(t *testing.T) {
	prev := coretenant.Enabled()
	coretenant.SetEnabled(false)
	t.Cleanup(func() { coretenant.SetEnabled(prev) })

	tmpDir := t.TempDir()
	policyPath := filepath.Join(tmpDir, "policy.csv")
	policy := "p, admin, /api/v1/users, GET\n"
	if err := os.WriteFile(policyPath, []byte(policy), 0o600); err != nil {
		t.Fatalf("write policy file: %v", err)
	}

	service, err := NewPermissionService(Config{
		Enabled:    true,
		ModelPath:  filepath.Clean(filepath.Join("..", "model", "rbac.conf")),
		PolicyPath: policyPath,
	})
	if err != nil {
		t.Fatalf("NewPermissionService: %v", err)
	}

	allowed, err := service.EnforceClaims(&coreauthjwt.Claims{
		Identity: coreauthjwt.Identity{
			TenantID: "tenant-b",
			Username: "alice",
			Roles:    []string{"admin"},
		},
	}, "/api/v1/users", "GET")
	if err != nil {
		t.Fatalf("EnforceClaims returned error: %v", err)
	}
	if !allowed {
		t.Fatal("expected role-only policy to allow access when tenant is disabled")
	}
}

func TestPermissionServiceDBSourceSeedsAndEnforcesPolicies(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "casbin.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}

	policyPath := filepath.Join(tmpDir, "policy.csv")
	if err := os.WriteFile(policyPath, []byte("p, admin, /api/v1/users, GET\n"), 0o600); err != nil {
		t.Fatalf("write policy file: %v", err)
	}

	service, err := NewPermissionService(Config{
		Enabled:    true,
		Source:     "db",
		DB:         db,
		ModelPath:  filepath.Clean(filepath.Join("..", "model", "rbac.conf")),
		PolicyPath: policyPath,
	})
	if err != nil {
		t.Fatalf("NewPermissionService(db): %v", err)
	}

	allowed, err := service.EnforceClaims(&coreauthjwt.Claims{Identity: coreauthjwt.Identity{Roles: []string{"admin"}}}, "/api/v1/users", "GET")
	if err != nil {
		t.Fatalf("EnforceClaims returned error: %v", err)
	}
	if !allowed {
		t.Fatal("expected db-backed casbin policy to allow access")
	}

	var policyCount int64
	if err := db.Table("casbin_rule").Count(&policyCount).Error; err != nil {
		t.Fatalf("count casbin_rule records: %v", err)
	}
	if policyCount == 0 {
		t.Fatal("expected policy rows to be seeded into casbin_rule")
	}

	var modelCount int64
	if err := db.Table("casbin_model").Count(&modelCount).Error; err != nil {
		t.Fatalf("count casbin_model records: %v", err)
	}
	if modelCount == 0 {
		t.Fatal("expected casbin model to be seeded into casbin_model")
	}
}
