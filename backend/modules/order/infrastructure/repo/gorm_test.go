package repo

import (
	"context"
	"strings"
	"testing"

	"goadmin/modules/order/domain/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestGormRepositoryCreateAssignsID(t *testing.T) {
	t.Parallel()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	if err := db.AutoMigrate(&model.Order{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}

	repo, err := NewGormRepository(db)
	if err != nil {
		t.Fatalf("new repo: %v", err)
	}

	created, err := repo.Create(context.Background(), &model.Order{OrderNo: "ORD-001"})
	if err != nil {
		t.Fatalf("create order: %v", err)
	}
	if strings.TrimSpace(created.Id) == "" {
		t.Fatalf("created order id is empty")
	}
	if !strings.HasPrefix(created.Id, "order-") {
		t.Fatalf("created order id = %q, want prefix order-", created.Id)
	}
}
