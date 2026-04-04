package http

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	downloadapp "goadmin/codegen/application/download"
	installapp "goadmin/codegen/application/install"
	cli "goadmin/codegen/driver/cli"
	"goadmin/core/response"
	menuservice "goadmin/modules/menu/application/service"
	menumodel "goadmin/modules/menu/domain/model"
	menurepopkg "goadmin/modules/menu/infrastructure/repo"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type fakeContext struct {
	requestContext context.Context
	params         map[string]string
	payload        any
	status         int
	jsonBody       any
	attachmentPath string
	attachmentName string
	headers        map[string]string
	values         map[string]any
}

func (c *fakeContext) RequestContext() context.Context {
	if c.requestContext == nil {
		return context.Background()
	}
	return c.requestContext
}

func (c *fakeContext) SetRequestContext(ctx context.Context) { c.requestContext = ctx }
func (c *fakeContext) Method() string                        { return "POST" }
func (c *fakeContext) Path() string                          { return "/api/v1/codegen/dsl/generate-download" }
func (c *fakeContext) Header(string) string                  { return "" }
func (c *fakeContext) SetHeader(key, value string) {
	if c.headers == nil {
		c.headers = make(map[string]string)
	}
	c.headers[key] = value
}
func (c *fakeContext) Param(key string) string {
	if c.params == nil {
		return ""
	}
	return c.params[key]
}
func (c *fakeContext) Query(string) string { return "" }
func (c *fakeContext) Set(key string, value any) {
	if c.values == nil {
		c.values = make(map[string]any)
	}
	c.values[key] = value
}
func (c *fakeContext) Get(key string) (any, bool) {
	if c.values == nil {
		return nil, false
	}
	value, ok := c.values[key]
	return value, ok
}
func (c *fakeContext) ShouldBindJSON(v any) error {
	data, err := json.Marshal(c.payload)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}
func (c *fakeContext) ShouldBindQuery(any) error { return nil }
func (c *fakeContext) BindJSON(v any) error      { return c.ShouldBindJSON(v) }
func (c *fakeContext) JSON(status int, payload any) {
	c.status = status
	c.jsonBody = payload
}
func (c *fakeContext) FileAttachment(path, name string) {
	c.attachmentPath = path
	c.attachmentName = name
}
func (c *fakeContext) AbortWithStatusJSON(status int, payload any) {
	c.status = status
	c.jsonBody = payload
}

func TestHandlerGenerateDownloadAndArtifact(t *testing.T) {
	t.Parallel()

	handler := NewHandler(Dependencies{
		ProjectRoot:     t.TempDir(),
		ArtifactEnabled: true,
		ArtifactBaseDir: t.TempDir(),
		ArtifactTTL:     time.Hour,
	})
	generateCtx := &fakeContext{
		payload: GenerateDownloadRequest{
			DSL: strings.TrimSpace(`
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
permissions:
  - inventory:view
`),
			PackageName: "inventory-module",
		},
	}

	handler.GenerateDownload(generateCtx)
	if generateCtx.status != 200 {
		t.Fatalf("GenerateDownload status = %d, want 200, body=%#v", generateCtx.status, generateCtx.jsonBody)
	}
	envelope, ok := generateCtx.jsonBody.(response.Envelope)
	if !ok {
		t.Fatalf("GenerateDownload body type = %T, want response.Envelope", generateCtx.jsonBody)
	}
	artifact, ok := envelope.Data.(downloadapp.ArtifactInfo)
	if !ok {
		t.Fatalf("GenerateDownload data type = %T, want download.ArtifactInfo", envelope.Data)
	}
	if artifact.TaskID == "" {
		t.Fatal("expected task id")
	}
	if artifact.DownloadURL == "" {
		t.Fatal("expected download url")
	}

	downloadCtx := &fakeContext{params: map[string]string{"taskID": artifact.TaskID}}
	handler.DownloadArtifact(downloadCtx)
	if downloadCtx.attachmentPath == "" {
		t.Fatal("expected attachment path")
	}
	if downloadCtx.attachmentName == "" {
		t.Fatal("expected attachment name")
	}
	if got := downloadCtx.headers["Cache-Control"]; got != "private, max-age=300" {
		t.Fatalf("Cache-Control = %q, want private, max-age=300", got)
	}
}

func TestHandlerPreviewDatabase(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	dbPath := filepath.Join(t.TempDir(), "codegen.db")
	db := openHTTPIntegrationSQLiteDB(t, dbPath)
	if err := db.Exec(`CREATE TABLE books (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL
	);`).Error; err != nil {
		t.Fatalf("create table: %v", err)
	}

	handler := NewHandler(Dependencies{ProjectRoot: root})
	ctx := &fakeContext{payload: DatabaseRequest{
		Driver:   "sqlite",
		DSN:      dbPath,
		Database: "codegen",
		Tables:   []string{"books"},
	}}

	handler.PreviewDatabase(ctx)
	if ctx.status != 200 {
		t.Fatalf("PreviewDatabase status = %d, want 200, body=%#v", ctx.status, ctx.jsonBody)
	}
	envelope, ok := ctx.jsonBody.(response.Envelope)
	if !ok {
		t.Fatalf("PreviewDatabase body type = %T, want response.Envelope", ctx.jsonBody)
	}
	report, ok := envelope.Data.(cli.DatabasePreviewReport)
	if !ok {
		t.Fatalf("PreviewDatabase data type = %T, want cli.DatabasePreviewReport", envelope.Data)
	}
	if !report.DryRun {
		t.Fatal("expected dry-run report")
	}
	if len(report.Resources) != 1 {
		t.Fatalf("expected 1 preview resource, got %d", len(report.Resources))
	}
	if report.Resources[0].Name != "book" {
		t.Fatalf("expected preview resource name book, got %q", report.Resources[0].Name)
	}
	if len(report.Files) == 0 {
		t.Fatal("expected file plan entries")
	}
}

func TestHandlerGenerateDatabaseDownloadAndArtifact(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	dbPath := filepath.Join(t.TempDir(), "codegen.db")
	db := openHTTPIntegrationSQLiteDB(t, dbPath)
	if err := db.Exec(`CREATE TABLE books (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL
	);`).Error; err != nil {
		t.Fatalf("create table: %v", err)
	}

	handler := NewHandler(Dependencies{
		ProjectRoot:     root,
		ArtifactEnabled: true,
		ArtifactBaseDir: t.TempDir(),
		ArtifactTTL:     time.Hour,
	})
	generateCtx := &fakeContext{payload: DatabaseRequest{
		Driver:   "sqlite",
		DSN:      dbPath,
		Database: "codegen",
		Tables:   []string{"books"},
	}}

	handler.GenerateDatabaseDownload(generateCtx)
	if generateCtx.status != 200 {
		t.Fatalf("GenerateDatabaseDownload status = %d, want 200, body=%#v", generateCtx.status, generateCtx.jsonBody)
	}
	envelope, ok := generateCtx.jsonBody.(response.Envelope)
	if !ok {
		t.Fatalf("GenerateDatabaseDownload body type = %T, want response.Envelope", generateCtx.jsonBody)
	}
	artifact, ok := envelope.Data.(downloadapp.ArtifactInfo)
	if !ok {
		t.Fatalf("GenerateDatabaseDownload data type = %T, want download.ArtifactInfo", envelope.Data)
	}
	if artifact.TaskID == "" {
		t.Fatal("expected task id")
	}
	if artifact.DownloadURL == "" {
		t.Fatal("expected download url")
	}

	downloadCtx := &fakeContext{params: map[string]string{"taskID": artifact.TaskID}}
	handler.DownloadArtifact(downloadCtx)
	if downloadCtx.attachmentPath == "" {
		t.Fatal("expected attachment path")
	}
	if downloadCtx.attachmentName == "" {
		t.Fatal("expected attachment name")
	}
	if got := downloadCtx.headers["Cache-Control"]; got != "private, max-age=300" {
		t.Fatalf("Cache-Control = %q, want private, max-age=300", got)
	}
	if strings.Contains(mustJSONString(t, envelope), dbPath) {
		t.Fatalf("HTTP generate-download response leaked DSN/path %q in JSON: %s", dbPath, mustJSONString(t, envelope))
	}
}

func TestHandlerInstallManifest(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	manifestDir := filepath.Join(root, "backend", "modules", "book")
	if err := os.MkdirAll(manifestDir, 0o755); err != nil {
		t.Fatalf("mkdir manifest dir: %v", err)
	}
	manifestPath := filepath.Join(manifestDir, "manifest.yaml")
	if err := os.WriteFile(manifestPath, []byte(strings.TrimSpace(`
name: book
version: v1
kind: crud
module: book
menus:
  - name: Books
    path: /books
    parent_path: /system
    component: Layout
    type: directory
    redirect: /books/list
    visible: true
    enabled: true
    sort: 1
  - name: List
    path: /books/list
    parent_path: /books
    component: view/book/index
    type: menu
    visible: true
    enabled: true
    sort: 2
`)), 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	dbPath := filepath.Join(t.TempDir(), "menus.db")
	db := openHTTPIntegrationSQLiteDB(t, dbPath)
	if err := menurepopkg.Migrate(db); err != nil {
		t.Fatalf("migrate menus: %v", err)
	}
	if err := menurepopkg.SeedDefaults(db); err != nil {
		t.Fatalf("seed default menus: %v", err)
	}
	menuRepo, err := menurepopkg.NewGormRepository(db)
	if err != nil {
		t.Fatalf("new menu repository: %v", err)
	}
	menuSvc, err := menuservice.New(menuRepo)
	if err != nil {
		t.Fatalf("new menu service: %v", err)
	}

	handler := NewHandler(Dependencies{ProjectRoot: root, MenuService: menuSvc})
	ctx := &fakeContext{payload: InstallManifestRequest{Module: "book"}}

	handler.InstallManifest(ctx)
	if ctx.status != 200 {
		t.Fatalf("InstallManifest status = %d, want 200, body=%#v", ctx.status, ctx.jsonBody)
	}
	envelope, ok := ctx.jsonBody.(response.Envelope)
	if !ok {
		t.Fatalf("InstallManifest body type = %T, want response.Envelope", ctx.jsonBody)
	}
	result, ok := envelope.Data.(installapp.InstallResult)
	if !ok {
		t.Fatalf("InstallManifest data type = %T, want install.InstallResult", envelope.Data)
	}
	if result.MenuTotal != 2 {
		t.Fatalf("menu_total = %d, want 2", result.MenuTotal)
	}
	if result.CreatedCount != 2 {
		t.Fatalf("created_count = %d, want 2", result.CreatedCount)
	}

	tree, err := menuSvc.Tree(context.Background())
	if err != nil {
		t.Fatalf("load menu tree: %v", err)
	}
	booksMenu, ok := findMenuByPath(tree, "/books")
	if !ok {
		t.Fatal("expected /books menu to exist after install")
	}
	if booksMenu.ParentID == "" {
		t.Fatal("expected /books menu to have a parent id from /system")
	}
	systemMenu, ok := findMenuByPath(tree, "/system")
	if !ok {
		t.Fatal("expected /system menu to exist")
	}
	if booksMenu.ParentID != systemMenu.ID {
		t.Fatalf("books menu parent_id = %q, want %q", booksMenu.ParentID, systemMenu.ID)
	}
	listMenu, ok := findMenuByPath(tree, "/books/list")
	if !ok {
		t.Fatal("expected /books/list menu to exist after install")
	}
	if listMenu.ParentID != booksMenu.ID {
		t.Fatalf("list menu parent_id = %q, want %q", listMenu.ParentID, booksMenu.ID)
	}
}

func TestHandlerGenerateDatabase(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	dbPath := filepath.Join(t.TempDir(), "codegen.db")
	db := openHTTPIntegrationSQLiteDB(t, dbPath)
	if err := db.Exec(`CREATE TABLE books (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL
	);`).Error; err != nil {
		t.Fatalf("create table: %v", err)
	}

	handler := NewHandler(Dependencies{ProjectRoot: root})
	ctx := &fakeContext{payload: DatabaseRequest{
		Driver:   "sqlite",
		DSN:      dbPath,
		Database: "codegen",
		Tables:   []string{"books"},
	}}

	handler.GenerateDatabase(ctx)
	if ctx.status != 200 {
		t.Fatalf("GenerateDatabase status = %d, want 200, body=%#v", ctx.status, ctx.jsonBody)
	}
	envelope, ok := ctx.jsonBody.(response.Envelope)
	if !ok {
		t.Fatalf("GenerateDatabase body type = %T, want response.Envelope", ctx.jsonBody)
	}
	report, ok := envelope.Data.(cli.DatabasePreviewReport)
	if !ok {
		t.Fatalf("GenerateDatabase data type = %T, want cli.DatabasePreviewReport", envelope.Data)
	}
	if report.DryRun {
		t.Fatal("expected generate report, got dry-run")
	}
	assertExists(t, filepath.Join(root, "backend", "modules", "book", "module.go"))
	assertExists(t, filepath.Join(root, "web", "src", "views", "book", "index.vue"))
	if !strings.Contains(strings.Join(report.Messages, " "), "generated 1 resource(s)") {
		t.Fatalf("GenerateDatabase report missing generated message: %+v", report.Messages)
	}
	if report.Audit.RecordedAt == "" {
		t.Fatal("expected audit record timestamp")
	}
	if len(report.Audit.Steps) == 0 {
		t.Fatal("expected audit steps")
	}
	if report.Audit.Output.FileCount != len(report.Files) {
		t.Fatalf("audit file count = %d, want %d", report.Audit.Output.FileCount, len(report.Files))
	}
	if strings.Contains(mustJSONString(t, envelope), dbPath) {
		t.Fatalf("HTTP generate response leaked DSN/path %q in JSON: %s", dbPath, mustJSONString(t, envelope))
	}
}

func TestHandlerGenerateDatabaseValidation(t *testing.T) {
	t.Parallel()

	handler := NewHandler(Dependencies{ProjectRoot: t.TempDir()})
	ctx := &fakeContext{payload: DatabaseRequest{DSN: "sqlite.db", Database: "codegen"}}

	handler.GenerateDatabase(ctx)
	if ctx.status != 400 {
		t.Fatalf("GenerateDatabase status = %d, want 400", ctx.status)
	}
	envelope, ok := ctx.jsonBody.(response.Envelope)
	if !ok {
		t.Fatalf("GenerateDatabase body type = %T, want response.Envelope", ctx.jsonBody)
	}
	if !strings.Contains(strings.ToLower(envelope.Message), "database driver is required") {
		t.Fatalf("GenerateDatabase message = %q, want database driver is required", envelope.Message)
	}
}

func openHTTPIntegrationSQLiteDB(t *testing.T, path string) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{Logger: gormlogger.Default.LogMode(gormlogger.Silent)})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	return db
}

func assertExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected %s to exist: %v", path, err)
	}
}

func mustJSONString(t *testing.T, value any) string {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal json: %v", err)
	}
	return string(data)
}

func findMenuByPath(items []menumodel.Menu, targetPath string) (menumodel.Menu, bool) {
	for _, item := range items {
		if strings.EqualFold(strings.TrimSpace(item.Path), strings.TrimSpace(targetPath)) {
			return item, true
		}
		if len(item.Children) == 0 {
			continue
		}
		if found, ok := findMenuByPath(item.Children, targetPath); ok {
			return found, true
		}
	}
	return menumodel.Menu{}, false
}
