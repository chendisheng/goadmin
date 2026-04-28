package gorm

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"goadmin/modules/upload/domain/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestRepositoryDefaultStorageDriver(t *testing.T) {
	t.Parallel()

	db, err := gorm.Open(sqlite.Open(testRepositorySQLiteDSN(t.Name())), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	repo, err := New(db)
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}

	driver, err := repo.DefaultStorageDriver(context.Background(), "local")
	if err != nil {
		t.Fatalf("default storage driver: %v", err)
	}
	if driver != "local" {
		t.Fatalf("unexpected fallback driver %q", driver)
	}

	if err := repo.SetDefaultStorageDriver(context.Background(), "qiniu"); err != nil {
		t.Fatalf("set default storage driver: %v", err)
	}

	driver, err = repo.DefaultStorageDriver(context.Background(), "local")
	if err != nil {
		t.Fatalf("reload default storage driver: %v", err)
	}
	if driver != "qiniu" {
		t.Fatalf("unexpected stored driver %q", driver)
	}
}

func testRepositorySQLiteDSN(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		name = "default"
	}
	name = strings.NewReplacer(" ", "_", "/", "_", "\\", "_").Replace(name)
	return fmt.Sprintf("file:repo-%s?mode=memory&cache=shared", name)
}

func TestRepositoryMigrateCreatesStorageSettingTable(t *testing.T) {
	t.Parallel()

	db, err := gorm.Open(sqlite.Open(testRepositorySQLiteDSN(t.Name())), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if !db.Migrator().HasTable(&model.StorageSetting{}) {
		t.Fatal("expected upload_storage_setting table to exist")
	}
}
