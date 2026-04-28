package inspector

import (
	"path/filepath"
	"sort"
	"testing"

	"goadmin/codegen/schema/database"
	dbschema "goadmin/codegen/schema/database"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func TestParseEnumCommentSupportsExplicitFormats(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		text string
	}{
		{name: "semicolon ui and enum", text: "分类|ui=select;enum=tech=技术,novel=小说,history=历史,other=其他"},
		{name: "bare enum list", text: "分类|tech=技术,novel=小说,history=历史,other=其他"},
		{name: "parenthesized ui and enum", text: "分类|(ui:select, enum:tech=技术,novel=小说,history=历史,other=其他)"},
		{name: "enum only parenthesized", text: "分类|(enum:tech=技术,novel=小说,history=历史,other=其他)"},
		{name: "parenthesized ui equals enum", text: "分类|(ui=select;enum=tech=技术,novel=小说,history=历史,other=其他)"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			parsed := parseEnumComment(tc.text)
			if !parsed.OK {
				t.Fatalf("parseEnumComment(%q) returned OK=false, want true", tc.text)
			}
			if got, want := parsed.Source, "comment"; got != want {
				t.Fatalf("Source = %q, want %q", got, want)
			}
			if got, want := parsed.Display, "select"; got != want {
				t.Fatalf("Display = %q, want %q", got, want)
			}
			if got, want := len(parsed.Options), 4; got != want {
				t.Fatalf("Options len = %d, want %d", got, want)
			}
			if got, want := parsed.Values, []string{"tech", "novel", "history", "other"}; !sameStrings(got, want) {
				t.Fatalf("Values = %v, want %v", got, want)
			}
			if got, want := parsed.Options[0].Value, "tech"; got != want {
				t.Fatalf("first option value = %q, want %q", got, want)
			}
			if got, want := parsed.Options[0].Label, "技术"; got != want {
				t.Fatalf("first option label = %q, want %q", got, want)
			}
		})
	}
}

func TestApplyEnumCommentMetadataParsesExplicitUIType(t *testing.T) {
	t.Parallel()

	column := dbschema.Column{Comment: "订单状态|ui=radio;enum=draft=草稿,published=已发布,archived=已归档"}
	applyEnumCommentMetadata(&column)

	if got, want := column.UIType, "radio"; got != want {
		t.Fatalf("UIType = %q, want %q", got, want)
	}
	if got, want := column.EnumDisplay, "radio"; got != want {
		t.Fatalf("EnumDisplay = %q, want %q", got, want)
	}
	if got, want := len(column.EnumValues), 3; got != want {
		t.Fatalf("EnumValues len = %d, want %d", got, want)
	}
	if got, want := column.Metadata["ui_type"], "radio"; got != want {
		t.Fatalf("ui_type metadata = %#v, want %q", got, want)
	}
}

func TestApplyEnumCommentMetadataDefaultsEnumUITypeToSelect(t *testing.T) {
	t.Parallel()

	column := dbschema.Column{Comment: "订单状态|enum=draft=草稿,published=已发布,archived=已归档"}
	applyEnumCommentMetadata(&column)

	if got, want := column.UIType, "select"; got != want {
		t.Fatalf("UIType = %q, want %q", got, want)
	}
	if got, want := column.EnumDisplay, "select"; got != want {
		t.Fatalf("EnumDisplay = %q, want %q", got, want)
	}
}

func TestParseEnumCommentRejectsOrdinaryDescription(t *testing.T) {
	t.Parallel()

	for _, text := range []string{
		"分类|这是一个普通说明，不应被解析为枚举",
		"分类|tech,novel,history,other",
		"分类|ui=radio",
		"分类|(wrong:enum=tech=技术,novel=小说,history=历史,other=其他)",
	} {
		text := text
		t.Run(text, func(t *testing.T) {
			t.Parallel()

			parsed := parseEnumComment(text)
			if parsed.OK {
				t.Fatalf("parseEnumComment returned OK=true, want false: %#v", parsed)
			}
			if len(parsed.Options) != 0 || len(parsed.Values) != 0 {
				t.Fatalf("expected empty enum result, got %#v", parsed)
			}
		})
	}
}

func TestGormInspectorSQLite(t *testing.T) {
	t.Parallel()

	db := openSQLiteTestDB(t)
	mustExec(t, db, `CREATE TABLE authors (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL
	);`)
	mustExec(t, db, `CREATE TABLE books (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		author_id INTEGER NOT NULL,
		title TEXT NOT NULL,
		price NUMERIC DEFAULT 0.00,
		CONSTRAINT fk_books_author FOREIGN KEY(author_id) REFERENCES authors(id) ON UPDATE CASCADE ON DELETE CASCADE
	);`)
	mustExec(t, db, `CREATE INDEX idx_books_title ON books(title);`)
	mustExec(t, db, `CREATE UNIQUE INDEX idx_books_author_title ON books(author_id, title);`)

	inspector := NewGormInspector(db)

	tables, err := inspector.InspectTables()
	if err != nil {
		t.Fatalf("InspectTables returned error: %v", err)
	}
	if len(tables) != 2 {
		t.Fatalf("expected 2 tables, got %d", len(tables))
	}
	if tables[0].Name != "authors" || tables[1].Name != "books" {
		t.Fatalf("unexpected tables order: %#v", []string{tables[0].Name, tables[1].Name})
	}
	if tables[1].Metadata["driver"] != database.DriverKindSQLite {
		t.Fatalf("expected sqlite driver metadata, got %#v", tables[1].Metadata)
	}

	books := tables[1]
	assertStringSliceContains(t, books.PrimaryKeys, "id")
	assertIndexPresent(t, books.Indexes, "idx_books_title", false, []string{"title"})
	assertIndexPresent(t, books.Indexes, "idx_books_author_title", true, []string{"author_id", "title"})
	assertForeignKeyPresent(t, books.ForeignKeys, "", "authors", []string{"author_id"}, []string{"id"})

	columns, err := inspector.InspectColumns("books")
	if err != nil {
		t.Fatalf("InspectColumns returned error: %v", err)
	}
	if len(columns) != 4 {
		t.Fatalf("expected 4 columns, got %d", len(columns))
	}

	byName := make(map[string]database.Column, len(columns))
	for _, column := range columns {
		byName[column.Name] = column
	}
	id := byName["id"]
	if !id.Primary {
		t.Fatalf("expected id to be primary: %#v", id)
	}
	if !id.AutoIncrement {
		t.Fatalf("expected id to be autoincrement: %#v", id)
	}
	if id.Type != "INTEGER" {
		t.Fatalf("expected id type INTEGER, got %q", id.Type)
	}

	authorID := byName["author_id"]
	if authorID.Name == "" || authorID.Type != "INTEGER" {
		t.Fatalf("unexpected author_id column: %#v", authorID)
	}

	title := byName["title"]
	if !title.Index {
		t.Fatalf("expected title to be indexed: %#v", title)
	}
	if title.Unique {
		t.Fatalf("did not expect title to be unique by itself: %#v", title)
	}

	price := byName["price"]
	if price.Default != "0.00" {
		t.Fatalf("expected price default 0.00, got %q", price.Default)
	}

	relations, err := inspector.InspectRelations("books")
	if err != nil {
		t.Fatalf("InspectRelations returned error: %v", err)
	}
	assertForeignKeyPresent(t, relations, "", "authors", []string{"author_id"}, []string{"id"})
}

func TestInspectTablesCarriesTableCommentFromInspectorContext(t *testing.T) {
	t.Parallel()

	db := openSQLiteTestDB(t)
	mustExec(t, db, `CREATE TABLE books (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL
	);`)

	inspector := NewGormInspector(db)
	if inspector == nil {
		t.Fatal("NewGormInspector returned nil")
	}
	if got := inspector.WithContext("goadmin", "public"); got == nil {
		t.Fatal("WithContext returned nil")
	}

	tables, err := inspector.WithContext("goadmin", "public").InspectTables()
	if err != nil {
		t.Fatalf("InspectTables returned error: %v", err)
	}
	if len(tables) != 1 {
		t.Fatalf("expected 1 table, got %d", len(tables))
	}
	if tables[0].Schema != "public" {
		t.Fatalf("expected schema public, got %q", tables[0].Schema)
	}
	if tables[0].Metadata["database"] != "goadmin" {
		t.Fatalf("expected database metadata goadmin, got %#v", tables[0].Metadata)
	}
	if tables[0].Comment != "" {
		t.Fatalf("expected sqlite table comment to be empty, got %q", tables[0].Comment)
	}
}

func TestApplyColumnCommentsPopulatesNonPrimaryFieldComments(t *testing.T) {
	t.Parallel()

	columns := []database.Column{
		{Name: "id", Primary: true},
		{Name: "title"},
		{Name: "status"},
	}
	applyColumnComments(columns, map[string]string{
		"id":     "订单ID",
		"title":  "标题",
		"status": "状态|enum=draft=草稿,published=已发布",
	})

	if got, want := columns[0].Comment, "订单ID"; got != want {
		t.Fatalf("primary column comment = %q, want %q", got, want)
	}
	if got, want := columns[1].Comment, "标题"; got != want {
		t.Fatalf("non-primary column comment = %q, want %q", got, want)
	}
	if got, want := columns[2].Comment, "状态|enum=draft=草稿,published=已发布"; got != want {
		t.Fatalf("enum column comment = %q, want %q", got, want)
	}
	if columns[2].UIType != "select" || columns[2].EnumKind == "" || len(columns[2].EnumValues) == 0 {
		t.Fatalf("expected enum metadata to be inferred from comment: %#v", columns[2])
	}
}

func openSQLiteTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dir := t.TempDir()
	dsn := filepath.Join(dir, "inspect.db")
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: gormlogger.Default.LogMode(gormlogger.Silent)})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	return db
}

func mustExec(t *testing.T, db *gorm.DB, sql string) {
	t.Helper()
	if err := db.Exec(sql).Error; err != nil {
		t.Fatalf("exec %q: %v", sql, err)
	}
}

func assertStringSliceContains(t *testing.T, values []string, want string) {
	t.Helper()
	for _, value := range values {
		if value == want {
			return
		}
	}
	t.Fatalf("expected %q in %v", want, values)
}

func assertIndexPresent(t *testing.T, indexes []database.Index, wantName string, wantUnique bool, wantColumns []string) {
	t.Helper()
	for _, index := range indexes {
		if index.Name != wantName {
			continue
		}
		if index.Unique != wantUnique {
			t.Fatalf("index %s unique mismatch: want %v got %v", wantName, wantUnique, index.Unique)
		}
		if !sameStrings(index.Columns, wantColumns) {
			t.Fatalf("index %s columns mismatch: want %v got %v", wantName, wantColumns, index.Columns)
		}
		return
	}
	t.Fatalf("index %s not found in %v", wantName, indexes)
}

func assertForeignKeyPresent(t *testing.T, foreignKeys []database.ForeignKey, wantName string, wantRefTable string, wantColumns []string, wantRefColumns []string) {
	t.Helper()
	for _, foreignKey := range foreignKeys {
		if wantName != "" && foreignKey.Name != wantName {
			continue
		}
		if foreignKey.RefTable != wantRefTable {
			t.Fatalf("foreign key %s ref table mismatch: want %s got %s", wantName, wantRefTable, foreignKey.RefTable)
		}
		if !sameStrings(foreignKey.Columns, wantColumns) {
			t.Fatalf("foreign key %s columns mismatch: want %v got %v", wantName, wantColumns, foreignKey.Columns)
		}
		if !sameStrings(foreignKey.RefColumns, wantRefColumns) {
			t.Fatalf("foreign key %s ref columns mismatch: want %v got %v", wantName, wantRefColumns, foreignKey.RefColumns)
		}
		return
	}
	t.Fatalf("foreign key %s not found in %v", wantName, foreignKeys)
}

func sameStrings(values []string, want []string) bool {
	if len(values) != len(want) {
		return false
	}
	got := append([]string(nil), values...)
	expected := append([]string(nil), want...)
	sort.Strings(got)
	sort.Strings(expected)
	for i := range got {
		if got[i] != expected[i] {
			return false
		}
	}
	return true
}
