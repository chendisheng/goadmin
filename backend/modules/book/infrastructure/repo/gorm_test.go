package repo

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"goadmin/modules/book/domain/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestGormRepositoryCreateAssignsID(t *testing.T) {
	t.Parallel()

	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "book-create.db")), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	if err := db.AutoMigrate(&model.Book{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}

	repo, err := NewGormRepository(db)
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}

	created, err := repo.Create(context.Background(), &model.Book{Title: "Go Admin", TenantId: "tenant-1"})
	if err != nil {
		t.Fatalf("create book: %v", err)
	}
	if strings.TrimSpace(created.Id) == "" {
		t.Fatalf("created book id is empty")
	}
	if !strings.HasPrefix(created.Id, "book-") {
		t.Fatalf("created book id = %q, want prefix book-", created.Id)
	}
}

func TestGormRepositoryListFiltersByKeyword(t *testing.T) {
	t.Parallel()

	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "book-list.db")), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	if err := db.AutoMigrate(&model.Book{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}

	repo, err := NewGormRepository(db)
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}

	now := time.Now().UTC()
	items := []model.Book{
		{Id: "book-1", Title: "Go Admin", Author: "Alice", Status: "published", CreatedAt: now, UpdatedAt: now},
		{Id: "book-2", Title: "Rust Handbook", Author: "Bob", Status: "draft", CreatedAt: now, UpdatedAt: now},
	}
	for _, item := range items {
		if err := db.Create(&item).Error; err != nil {
			t.Fatalf("seed book %q: %v", item.Id, err)
		}
	}

	filtered, total, err := repo.List(context.Background(), "go", 1, 10)
	if err != nil {
		t.Fatalf("list filtered books: %v", err)
	}
	if total != 1 {
		t.Fatalf("filtered total = %d, want 1", total)
	}
	if len(filtered) != 1 {
		t.Fatalf("filtered item count = %d, want 1", len(filtered))
	}
	if filtered[0].Id != "book-1" {
		t.Fatalf("filtered book id = %q, want book-1", filtered[0].Id)
	}

	allItems, allTotal, err := repo.List(context.Background(), "", 1, 10)
	if err != nil {
		t.Fatalf("list all books: %v", err)
	}
	if allTotal != 2 {
		t.Fatalf("unfiltered total = %d, want 2", allTotal)
	}
	if len(allItems) != 2 {
		t.Fatalf("unfiltered item count = %d, want 2", len(allItems))
	}
}
