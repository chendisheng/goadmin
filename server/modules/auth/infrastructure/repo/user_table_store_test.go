package repo

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"goadmin/core/config"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestUserTableStoreAuthenticateFromUserTable(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	createUserAuthTables(t, db)
	insertUserRow(t, db, "u1", "system", "alice", "Alice", "zh-CN", "sha256:"+sha256Hex("123456"), `["user"]`)
	insertRoleRow(t, db, "r1", "system", "user", `["menu-users"]`)
	insertMenuRow(t, db, "menu-users", "user:list")

	store := NewUserTableStore(db, nil)
	identity, err := store.Authenticate(context.Background(), "alice", "123456")
	if err != nil {
		t.Fatalf("Authenticate: %v", err)
	}
	if identity.Username != "alice" {
		t.Fatalf("expected username alice, got %q", identity.Username)
	}
	if got := len(identity.Roles); got != 1 || identity.Roles[0] != "user" {
		t.Fatalf("expected roles [user], got %#v", identity.Roles)
	}
	if got := len(identity.Permissions); got != 1 || identity.Permissions[0] != "user:list" {
		t.Fatalf("expected permissions [user:list], got %#v", identity.Permissions)
	}
}

func TestUserTableStoreAuthenticateAdminUsesWildcard(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	createUserAuthTables(t, db)
	insertUserRow(t, db, "u1", "system", "admin", "Admin", "zh-CN", "sha256:"+sha256Hex("123456"), `["admin"]`)

	store := NewUserTableStore(db, nil)
	identity, err := store.Authenticate(context.Background(), "admin", "123456")
	if err != nil {
		t.Fatalf("Authenticate: %v", err)
	}
	if got := len(identity.Permissions); got != 1 || identity.Permissions[0] != "*" {
		t.Fatalf("expected wildcard permissions, got %#v", identity.Permissions)
	}
}

func TestUserTableStoreFallsBackToBootstrapCredentials(t *testing.T) {
	t.Parallel()

	db := openTestDB(t)
	createUserAuthTables(t, db)
	fallback := NewBootstrapStore([]config.BootstrapUser{{Username: "bootstrap", PasswordHash: "sha256:" + sha256Hex("secret"), Roles: []string{"admin"}}})

	store := NewUserTableStore(db, fallback)
	identity, err := store.Authenticate(context.Background(), "bootstrap", "secret")
	if err != nil {
		t.Fatalf("Authenticate: %v", err)
	}
	if identity.Username != "bootstrap" {
		t.Fatalf("expected fallback username bootstrap, got %q", identity.Username)
	}
	if got := len(identity.Permissions); got != 1 || identity.Permissions[0] != "*" {
		t.Fatalf("expected fallback wildcard permissions, got %#v", identity.Permissions)
	}
}

func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file:auth_repo_test_"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	if err := db.Exec(`PRAGMA foreign_keys = ON`).Error; err != nil {
		t.Fatalf("enable sqlite foreign keys: %v", err)
	}
	return db
}

func createUserAuthTables(t *testing.T, db *gorm.DB) {
	t.Helper()

	statements := []string{
		`CREATE TABLE IF NOT EXISTS user (
			id TEXT PRIMARY KEY,
			tenant_id TEXT,
			username TEXT NOT NULL,
			display_name TEXT,
			language TEXT,
			password_hash TEXT,
			role_codes TEXT NOT NULL,
			created_at DATETIME,
			updated_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS role (
			id TEXT PRIMARY KEY,
			tenant_id TEXT,
			code TEXT NOT NULL,
			menu_ids TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS menu (
			id TEXT PRIMARY KEY,
			permission TEXT
		)`,
	}
	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatalf("exec schema statement: %v", err)
		}
	}
}

func insertUserRow(t *testing.T, db *gorm.DB, id, tenantID, username, displayName, language, passwordHash, roleCodes string) {
	t.Helper()

	if err := db.Exec(`INSERT INTO user (id, tenant_id, username, display_name, language, password_hash, role_codes, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`, id, tenantID, username, displayName, language, passwordHash, roleCodes).Error; err != nil {
		t.Fatalf("insert user row: %v", err)
	}
}

func insertRoleRow(t *testing.T, db *gorm.DB, id, tenantID, code, menuIDs string) {
	t.Helper()

	if err := db.Exec(`INSERT INTO role (id, tenant_id, code, menu_ids) VALUES (?, ?, ?, ?)`, id, tenantID, code, menuIDs).Error; err != nil {
		t.Fatalf("insert role row: %v", err)
	}
}

func insertMenuRow(t *testing.T, db *gorm.DB, id, permission string) {
	t.Helper()

	if err := db.Exec(`INSERT INTO menu (id, permission) VALUES (?, ?)`, id, permission).Error; err != nil {
		t.Fatalf("insert menu row: %v", err)
	}
}

func sha256Hex(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

var _ CredentialStore = (*UserTableStore)(nil)
var _ CredentialStore = (*BootstrapStore)(nil)
