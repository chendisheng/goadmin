package inspector

import (
	"testing"

	"goadmin/codegen/schema/database"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func TestServiceOpenUsesInjectedFactory(t *testing.T) {
	t.Parallel()

	called := false
	service := NewService(FactoryFunc(func(db *gorm.DB) Reader {
		called = true
		return stubReader{}
	}))

	reader := service.Open(nil, "db-name", "schema-name")
	if !called {
		t.Fatalf("expected injected factory to be called")
	}
	if reader == nil {
		t.Fatalf("expected reader")
	}
	if _, ok := reader.(stubReader); !ok {
		t.Fatalf("expected stubReader, got %T", reader)
	}
	if _, ok := reader.WithContext("db-2", "schema-2").(stubReader); !ok {
		t.Fatalf("expected stubReader after context rebinding, got %T", reader)
	}
}

func TestDefaultFactoryBuildsGormInspector(t *testing.T) {
	t.Parallel()

	db := openFactoryTestDB(t)
	service := NewService(nil)
	reader := service.Open(db, "inspect_db", "main")
	if reader == nil {
		t.Fatalf("expected reader")
	}
	if _, ok := reader.(*GormInspector); !ok {
		t.Fatalf("expected *GormInspector, got %T", reader)
	}
}

type stubReader struct{}

func (stubReader) InspectTables() ([]database.Table, error)               { return nil, nil }
func (stubReader) InspectColumns(string) ([]database.Column, error)       { return nil, nil }
func (stubReader) InspectRelations(string) ([]database.ForeignKey, error) { return nil, nil }
func (stubReader) WithContext(string, string) Reader                      { return stubReader{} }

func openFactoryTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: gormlogger.Default.LogMode(gormlogger.Silent)})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	return db
}
