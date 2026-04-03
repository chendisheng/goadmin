package cli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func TestRunGenerateDSL(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	dslPath := filepath.Join(root, "inventory.yaml")
	dsl := strings.TrimSpace(`
module: inventory
kind: business-module
framework:
  backend: gin
  frontend: vue3
entity:
  name: item
  fields:
    - name: id
      type: string
      primary: true
    - name: name
      type: string
      required: true
pages:
  - list
  - form
permissions:
  - inventory:view
  - inventory:edit
`)
	if err := os.WriteFile(dslPath, []byte(dsl), 0o644); err != nil {
		t.Fatalf("write dsl file: %v", err)
	}

	if err := Run(root, []string{"generate", "dsl", dslPath}); err != nil {
		t.Fatalf("Run(generate dsl) returned error: %v", err)
	}

	modulePath := filepath.Join(root, "backend", "modules", "inventory", "module.go")
	crudModelPath := filepath.Join(root, "backend", "modules", "item", "domain", "model", "item.go")
	frontendViewPath := filepath.Join(root, "backend", "web", "src", "views", "item", "index.vue")
	policyPath := filepath.Join(root, "backend", "core", "auth", "casbin", "adapter", "policy.csv")

	assertFileContains(t, modulePath, "package inventory")
	assertFileContains(t, modulePath, `const Name = "inventory"`)
	assertFileContains(t, crudModelPath, "type Item struct")
	assertFileContains(t, crudModelPath, `gorm:"column:name"`)
	assertFileExists(t, frontendViewPath)
	assertFileContains(t, policyPath, "p, admin, /api/v1/items, GET")
	assertFileContains(t, policyPath, "p, admin, /api/v1/items/:id, DELETE")
}

func TestExecuteDSLDocumentDryRun(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	dsl := strings.TrimSpace(`
version: v1
resources:
  - kind: frontend-page
    name: codegen-console
    module: system
    pages:
      - name: console
        path: /system/codegen
        component: system/codegen/index
`)
	report, err := ExecuteDSLDocument(root, []byte(dsl), false, true)
	if err != nil {
		t.Fatalf("ExecuteDSLDocument(dry-run) returned error: %v", err)
	}
	if !report.DryRun {
		t.Fatalf("expected dry-run report")
	}
	if len(report.Items) != 1 {
		t.Fatalf("expected 1 preview item, got %d", len(report.Items))
	}
	if len(report.Items[0].Actions) == 0 {
		t.Fatalf("expected preview actions")
	}
	hasPageAction := false
	for _, action := range report.Items[0].Actions {
		if strings.Contains(action, "generate page") {
			hasPageAction = true
			break
		}
	}
	if !hasPageAction {
		t.Fatalf("unexpected preview action: %v", report.Items[0].Actions)
	}
	if _, err := os.Stat(filepath.Join(root, "backend", "web", "src", "views", "system", "codegen", "index.vue")); !os.IsNotExist(err) {
		t.Fatalf("dry-run should not create output files, got err=%v", err)
	}
}

func TestRunGenerateDBPreview(t *testing.T) {
	root := t.TempDir()
	dbPath := filepath.Join(t.TempDir(), "codegen.db")
	db := openCLIIntegrationSQLiteDB(t, dbPath)
	if err := db.Exec(`CREATE TABLE books (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL
	);`).Error; err != nil {
		t.Fatalf("create table: %v", err)
	}

	output, err := captureCLIStdout(t, func() error {
		return Run(root, []string{
			"generate",
			"db",
			"preview",
			"--driver", "sqlite",
			"--dsn", dbPath,
			"--database", "codegen",
			"--table", "books",
		})
	})
	if err != nil {
		t.Fatalf("Run(generate db preview) returned error: %v", err)
	}
	if !strings.Contains(output, "dsl dry-run: no files will be written") {
		t.Fatalf("preview output missing dry-run message:\n%s", output)
	}
	if !strings.Contains(output, "resource[0]") {
		t.Fatalf("preview output missing resource index:\n%s", output)
	}
	if !strings.Contains(output, "kind=crud") {
		t.Fatalf("preview output missing kind=crud:\n%s", output)
	}
	if !strings.Contains(output, "name=book") {
		t.Fatalf("preview output missing name=book:\n%s", output)
	}
}

func assertFileContains(t *testing.T, path string, want string) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if !strings.Contains(string(content), want) {
		t.Fatalf("%s does not contain %q\ncontent:\n%s", path, want, string(content))
	}
}

func assertFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
}

func openCLIIntegrationSQLiteDB(t *testing.T, path string) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{Logger: gormlogger.Default.LogMode(gormlogger.Silent)})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	return db
}

func captureCLIStdout(t *testing.T, fn func() error) (string, error) {
	t.Helper()
	original := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe stdout: %v", err)
	}
	os.Stdout = w
	defer func() { os.Stdout = original }()

	outputCh := make(chan string, 1)
	errCh := make(chan error, 1)
	go func() {
		var buf bytes.Buffer
		_, copyErr := io.Copy(&buf, r)
		if copyErr != nil {
			errCh <- copyErr
			return
		}
		outputCh <- buf.String()
	}()

	runErr := fn()
	_ = w.Close()
	output := <-outputCh
	select {
	case copyErr := <-errCh:
		return output, copyErr
	default:
	}
	return output, runErr
}
