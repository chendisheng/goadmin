package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	casbinadapter "goadmin/core/auth/casbin/adapter"
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
	if policyCount != 0 {
		t.Fatalf("expected startup to stay read-only, got %d policy rows", policyCount)
	}

	var modelCount int64
	if err := db.Table("casbin_model").Count(&modelCount).Error; err != nil {
		t.Fatalf("count casbin_model records: %v", err)
	}
	if modelCount == 0 {
		t.Fatal("expected casbin model to be seeded into casbin_model")
	}
}

func TestPermissionServiceDBSourceSeedsDefaultUploadPolicies(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "casbin.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}

	policyPath := filepath.Join(tmpDir, "policy.csv")
	if err := os.WriteFile(policyPath, nil, 0o600); err != nil {
		t.Fatalf("write empty policy file: %v", err)
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

	cases := []struct {
		name   string
		path   string
		method string
	}{
		{name: "list", path: "/api/v1/uploads/files", method: "GET"},
		{name: "upload", path: "/api/v1/uploads/files", method: "POST"},
		{name: "download", path: "/api/v1/uploads/files/:id/download", method: "GET"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			allowed, err := service.EnforceClaims(&coreauthjwt.Claims{Identity: coreauthjwt.Identity{Roles: []string{"admin"}}}, tc.path, tc.method)
			if err != nil {
				t.Fatalf("EnforceClaims returned error: %v", err)
			}
			if !allowed {
				t.Fatal("expected db-backed default policy to allow access")
			}
		})
	}

	var policyCount int64
	if err := db.Table("casbin_rule").Where("ptype = ?", "p").Count(&policyCount).Error; err != nil {
		t.Fatalf("count casbin_rule records: %v", err)
	}
	if policyCount != 0 {
		t.Fatalf("expected startup to keep built-in policies in memory only, got %d rows", policyCount)
	}
}

func TestPermissionServiceDBSourceMergesMissingPoliciesFromCSV(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "casbin.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}

	store, err := casbinadapter.NewGormStore(db)
	if err != nil {
		t.Fatalf("NewGormStore: %v", err)
	}
	if err := casbinadapter.Migrate(db); err != nil {
		t.Fatalf("migrate casbin tables: %v", err)
	}
	if err := store.SavePolicies(context.Background(), []casbinadapter.Rule{{Subject: "admin", Object: "/api/v1/users", Action: "GET"}}); err != nil {
		t.Fatalf("seed existing DB policy: %v", err)
	}

	policyPath := filepath.Join(tmpDir, "policy.csv")
	policy := "p, admin, /api/v1/users, GET\np, admin, /api/v1/codegen/delete/preview, POST\n"
	if err := os.WriteFile(policyPath, []byte(policy), 0o600); err != nil {
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

	allowed, err := service.EnforceClaims(&coreauthjwt.Claims{Identity: coreauthjwt.Identity{Roles: []string{"admin"}}}, "/api/v1/codegen/delete/preview", "POST")
	if err != nil {
		t.Fatalf("EnforceClaims returned error: %v", err)
	}
	if !allowed {
		t.Fatal("expected merged db-backed casbin policy to allow codegen delete preview")
	}

	var policyCount int64
	if err := db.Table("casbin_rule").Where("ptype = ?", "p").Count(&policyCount).Error; err != nil {
		t.Fatalf("count casbin_rule records: %v", err)
	}
	if policyCount != 1 {
		t.Fatalf("expected startup to preserve existing policy rows only, got %d", policyCount)
	}
}

func TestGormStoreSavePoliciesPreservesExistingRulesAndAddsMissingOnes(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "casbin.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}

	store, err := casbinadapter.NewGormStore(db)
	if err != nil {
		t.Fatalf("NewGormStore: %v", err)
	}
	if err := casbinadapter.Migrate(db); err != nil {
		t.Fatalf("migrate casbin tables: %v", err)
	}

	seeded := []casbinadapter.Rule{{Subject: "admin", Object: "/api/v1/users", Action: "GET"}}
	if err := store.SavePolicies(context.Background(), seeded); err != nil {
		t.Fatalf("seed existing policy: %v", err)
	}

	if err := store.SavePolicies(context.Background(), []casbinadapter.Rule{
		{Subject: "admin", Object: "/api/v1/users", Action: "GET"},
		{Subject: "admin", Object: "/api/v1/uploads/files", Action: "GET"},
	}); err != nil {
		t.Fatalf("save merged policies: %v", err)
	}

	var policyCount int64
	if err := db.Table("casbin_rule").Where("ptype = ?", "p").Count(&policyCount).Error; err != nil {
		t.Fatalf("count casbin_rule records: %v", err)
	}
	if policyCount != 2 {
		t.Fatalf("expected exactly 2 policy rows after incremental sync, got %d", policyCount)
	}

	var uploadCount int64
	if err := db.Table("casbin_rule").Where("ptype = ? AND v1 = ?", "p", "/api/v1/uploads/files").Count(&uploadCount).Error; err != nil {
		t.Fatalf("count upload policy records: %v", err)
	}
	if uploadCount != 1 {
		t.Fatalf("expected upload policy to be inserted once, got %d", uploadCount)
	}
}

func TestPermissionServiceDefaultPolicyAllowsCodegenDeletePreview(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	policyPath := filepath.Join(tmpDir, "policy.csv")
	if err := os.WriteFile(policyPath, nil, 0o600); err != nil {
		t.Fatalf("write empty policy file: %v", err)
	}

	service, err := NewPermissionService(Config{
		Enabled:    true,
		ModelPath:  filepath.Clean(filepath.Join("..", "model", "rbac.conf")),
		PolicyPath: policyPath,
	})
	if err != nil {
		t.Fatalf("NewPermissionService: %v", err)
	}

	allowed, err := service.EnforceClaims(&coreauthjwt.Claims{Identity: coreauthjwt.Identity{Roles: []string{"admin"}}}, "/api/v1/codegen/delete/preview", "POST")
	if err != nil {
		t.Fatalf("EnforceClaims returned error: %v", err)
	}
	if !allowed {
		t.Fatal("expected default policy to allow codegen delete preview")
	}
}

func TestPermissionServiceDefaultPolicyAllowsUploadList(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	policyPath := filepath.Join(tmpDir, "policy.csv")
	if err := os.WriteFile(policyPath, nil, 0o600); err != nil {
		t.Fatalf("write empty policy file: %v", err)
	}

	service, err := NewPermissionService(Config{
		Enabled:    true,
		ModelPath:  filepath.Clean(filepath.Join("..", "model", "rbac.conf")),
		PolicyPath: policyPath,
	})
	if err != nil {
		t.Fatalf("NewPermissionService: %v", err)
	}

	allowed, err := service.EnforceClaims(&coreauthjwt.Claims{Identity: coreauthjwt.Identity{Roles: []string{"admin"}}}, "/api/v1/uploads/files", "GET")
	if err != nil {
		t.Fatalf("EnforceClaims returned error: %v", err)
	}
	if !allowed {
		t.Fatal("expected default policy to allow upload file listing")
	}
}

func TestPermissionServiceDefaultPolicyAllowsUploadModuleRoutes(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	policyPath := filepath.Join(tmpDir, "policy.csv")
	if err := os.WriteFile(policyPath, nil, 0o600); err != nil {
		t.Fatalf("write empty policy file: %v", err)
	}

	service, err := NewPermissionService(Config{
		Enabled:    true,
		ModelPath:  filepath.Clean(filepath.Join("..", "model", "rbac.conf")),
		PolicyPath: policyPath,
	})
	if err != nil {
		t.Fatalf("NewPermissionService: %v", err)
	}

	cases := []struct {
		name   string
		path   string
		method string
	}{
		{name: "list", path: "/api/v1/uploads/files", method: "GET"},
		{name: "detail", path: "/api/v1/uploads/files/:id", method: "GET"},
		{name: "upload", path: "/api/v1/uploads/files", method: "POST"},
		{name: "delete", path: "/api/v1/uploads/files/:id", method: "DELETE"},
		{name: "download", path: "/api/v1/uploads/files/:id/download", method: "GET"},
		{name: "preview", path: "/api/v1/uploads/files/:id/preview", method: "GET"},
		{name: "bind", path: "/api/v1/uploads/files/:id/bind", method: "POST"},
		{name: "unbind", path: "/api/v1/uploads/files/:id/bind", method: "DELETE"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			allowed, err := service.EnforceClaims(&coreauthjwt.Claims{Identity: coreauthjwt.Identity{Roles: []string{"admin"}}}, tc.path, tc.method)
			if err != nil {
				t.Fatalf("EnforceClaims returned error: %v", err)
			}
			if !allowed {
				t.Fatalf("expected default policy to allow %s %s", tc.method, tc.path)
			}
		})
	}
}
