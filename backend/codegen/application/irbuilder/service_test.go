package irbuilder

import (
	"path/filepath"
	"testing"

	insp "goadmin/codegen/infrastructure/inspector"
	irmodel "goadmin/codegen/model/ir"
	dbschema "goadmin/codegen/schema/database"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func TestBuildFromDatabaseUsesInspectorService(t *testing.T) {
	t.Parallel()

	stub := &stubReader{
		tables: []dbschema.Table{{
			Name:        "books",
			Schema:      "main",
			PrimaryKeys: []string{"id"},
			Columns: []dbschema.Column{
				{Name: "id", Type: "INTEGER", Primary: true, AutoIncrement: true},
				{Name: "title", Type: "TEXT", Nullable: false, Index: true},
				{Name: "author_id", Type: "INTEGER", Nullable: false},
			},
			ForeignKeys: []dbschema.ForeignKey{{
				Name:       "fk_books_author",
				Columns:    []string{"author_id"},
				RefTable:   "authors",
				RefColumns: []string{"id"},
			}},
		}},
	}
	called := false
	factory := insp.FactoryFunc(func(db *gorm.DB) insp.Reader {
		called = true
		return stub
	})
	service := NewService(Dependencies{InspectorService: insp.NewService(factory)})

	doc, err := service.BuildFromDatabase(nil, "library_db", "main")
	if err != nil {
		t.Fatalf("BuildFromDatabase returned error: %v", err)
	}
	if !called {
		t.Fatalf("expected inspector factory to be called")
	}
	if stub.database != "library_db" || stub.schema != "main" {
		t.Fatalf("expected context to be applied, got database=%q schema=%q", stub.database, stub.schema)
	}
	if doc.Version != defaultIRVersion {
		t.Fatalf("expected version %q, got %q", defaultIRVersion, doc.Version)
	}
	if len(doc.Resources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(doc.Resources))
	}
	resource := doc.Resources[0]
	if resource.Name != "Book" {
		t.Fatalf("expected resource name Book, got %q", resource.Name)
	}
	if resource.TableName != "books" {
		t.Fatalf("expected table name books, got %q", resource.TableName)
	}
	if resource.Metadata["schema"] != "main" {
		t.Fatalf("expected schema metadata main, got %#v", resource.Metadata)
	}
	if len(resource.Fields) != 3 {
		t.Fatalf("expected 3 fields, got %d", len(resource.Fields))
	}
	assertField(t, resource.Fields, "Id", "id", "int64")
	assertField(t, resource.Fields, "Title", "title", "string")
	assertField(t, resource.Fields, "AuthorId", "author_id", "int64")
	if len(resource.Relations) != 1 {
		t.Fatalf("expected 1 relation, got %d", len(resource.Relations))
	}
	if resource.Relations[0].RefTable != "authors" {
		t.Fatalf("unexpected relation: %#v", resource.Relations[0])
	}
}

func TestBuildFromDatabaseWithDefaultFactory(t *testing.T) {
	t.Parallel()

	db := openIRBuilderTestDB(t)
	mustExecIR(t, db, `CREATE TABLE authors (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL
	);`)
	mustExecIR(t, db, `CREATE TABLE books (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		author_id INTEGER NOT NULL,
		title TEXT NOT NULL,
		price NUMERIC DEFAULT 0.00,
		CONSTRAINT fk_books_author FOREIGN KEY(author_id) REFERENCES authors(id) ON UPDATE CASCADE ON DELETE CASCADE
	);`)
	mustExecIR(t, db, `CREATE INDEX idx_books_title ON books(title);`)

	service := NewService(Dependencies{})
	doc, err := service.BuildFromDatabase(db, "library_db", "main")
	if err != nil {
		t.Fatalf("BuildFromDatabase returned error: %v", err)
	}
	if len(doc.Resources) != 2 {
		t.Fatalf("expected 2 resources, got %d", len(doc.Resources))
	}
	books := doc.Resources[1]
	if books.Name != "Book" {
		t.Fatalf("expected Book resource, got %q", books.Name)
	}
	if books.Metadata["database"] != "library_db" {
		t.Fatalf("expected database metadata library_db, got %#v", books.Metadata)
	}
	assertField(t, books.Fields, "Price", "price", "int64")
}

func openIRBuilderTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dir := t.TempDir()
	dsn := filepath.Join(dir, "builder.db")
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: gormlogger.Default.LogMode(gormlogger.Silent)})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	return db
}

func mustExecIR(t *testing.T, db *gorm.DB, sql string) {
	t.Helper()
	if err := db.Exec(sql).Error; err != nil {
		t.Fatalf("exec %q: %v", sql, err)
	}
}

func assertField(t *testing.T, fields []irmodel.Field, wantName string, wantColumn string, wantGoType string) {
	t.Helper()
	for _, field := range fields {
		if field.Name != wantName {
			continue
		}
		if field.ColumnName != wantColumn {
			t.Fatalf("field %s column mismatch: want %s got %s", wantName, wantColumn, field.ColumnName)
		}
		if field.GoType != wantGoType {
			t.Fatalf("field %s go type mismatch: want %s got %s", wantName, wantGoType, field.GoType)
		}
		return
	}
	t.Fatalf("field %s not found", wantName)
}

type stubReader struct {
	database string
	schema   string
	tables   []dbschema.Table
}

func (s *stubReader) InspectTables() ([]dbschema.Table, error) {
	return append([]dbschema.Table(nil), s.tables...), nil
}

func (s *stubReader) InspectColumns(table string) ([]dbschema.Column, error) {
	for _, item := range s.tables {
		if item.Name == table {
			return append([]dbschema.Column(nil), item.Columns...), nil
		}
	}
	return nil, nil
}

func (s *stubReader) InspectRelations(table string) ([]dbschema.ForeignKey, error) {
	for _, item := range s.tables {
		if item.Name == table {
			return append([]dbschema.ForeignKey(nil), item.ForeignKeys...), nil
		}
	}
	return nil, nil
}

func (s *stubReader) WithContext(database string, schema string) insp.Reader {
	s.database = database
	s.schema = schema
	return s
}
