package repo

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestMigrateRenamesLegacyUsersTable(t *testing.T) {
	t.Parallel()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}

	if err := db.Exec(`CREATE TABLE users (
		id TEXT PRIMARY KEY,
		tenant_id TEXT NOT NULL,
		username TEXT NOT NULL,
		display_name TEXT,
		mobile TEXT,
		email TEXT,
		status TEXT NOT NULL,
		role_codes TEXT NOT NULL,
		password_hash TEXT,
		created_at DATETIME,
		updated_at DATETIME
	)`).Error; err != nil {
		t.Fatalf("create legacy users table: %v", err)
	}
	if err := db.Exec(`INSERT INTO users (id, tenant_id, username, display_name, mobile, email, status, role_codes, password_hash, created_at, updated_at) VALUES ('u1', 'system', 'admin', 'Admin', '', 'admin@goadmin.local', 'active', '["admin"]', '', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`).Error; err != nil {
		t.Fatalf("insert legacy row: %v", err)
	}

	if err := Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	if !db.Migrator().HasTable(&userRecord{}) {
		t.Fatalf("expected singular user table after migrate")
	}
	if db.Migrator().HasTable("users") {
		t.Fatalf("expected legacy users table to be renamed away")
	}

	var count int64
	if err := db.Model(&userRecord{}).Where("id = ?", "u1").Count(&count).Error; err != nil {
		t.Fatalf("count migrated row: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected migrated row count 1, got %d", count)
	}
}
