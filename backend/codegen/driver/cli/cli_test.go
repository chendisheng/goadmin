package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	deletionapp "goadmin/codegen/application/deletion"
	deletionmodel "goadmin/codegen/model/deletion"
	"goadmin/codegen/schema"
	casbinadapter "goadmin/core/auth/casbin/adapter"
	menucommand "goadmin/modules/menu/application/command"
	menuservice "goadmin/modules/menu/application/service"
	menuModel "goadmin/modules/menu/domain/model"
	menurepopkg "goadmin/modules/menu/infrastructure/repo"

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
	frontendViewPath := filepath.Join(root, "web", "src", "views", "item", "index.vue")
	policyPath := filepath.Join(root, "backend", "core", "auth", "casbin", "adapter", "policy.csv")

	assertFileContains(t, modulePath, "package inventory")
	assertFileContains(t, modulePath, `const Name = "inventory"`)
	assertFileContains(t, crudModelPath, "type Item struct")
	assertFileContains(t, crudModelPath, `gorm:"column:name;type:varchar(255);size:255"`)
	assertFileExists(t, frontendViewPath)
	assertFileContains(t, policyPath, "p, admin, /api/v1/items, GET")
	assertFileContains(t, policyPath, "p, admin, /api/v1/items/:id, DELETE")
}

func TestRunRemovePreview(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	deletionRoot := filepath.Join(root, "backend")
	createDeleteModuleFixture(t, root, "book", true, true)
	output, err := captureCLIStdout(t, func() error {
		return Run(root, []string{"remove", "preview", "book", "--kind", "crud", "--policy-store", "db"})
	})
	if err != nil {
		t.Fatalf("Run(remove preview) returned error: %v", err)
	}
	if !strings.Contains(output, "deletion preview report") {
		t.Fatalf("preview output missing report header:\n%s", output)
	}
	if !strings.Contains(output, "module: book") {
		t.Fatalf("preview output missing module summary:\n%s", output)
	}
	if !strings.Contains(output, "conflicts:") {
		t.Fatalf("preview output missing conflicts section:\n%s", output)
	}
	if !strings.Contains(output, "source files:") {
		t.Fatalf("preview output missing source files section:\n%s", output)
	}
	if !strings.Contains(output, "summary:") {
		t.Fatalf("preview output missing summary section:\n%s", output)
	}
	if _, err := os.Stat(filepath.Join(deletionRoot, "modules", "book", "module.go")); err != nil {
		t.Fatalf("fixture not created: %v", err)
	}
}

func TestRunRemovePreviewValidation(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := Run(root, []string{"remove", "preview"}); err == nil || !strings.Contains(err.Error(), "remove preview requires a module name") {
		t.Fatalf("Run(remove preview) validation error = %v, want module name required", err)
	}
}

func TestRunRemoveExecuteWithRuntimeDependencies(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	createDeleteModuleFixture(t, root, "book", true, false)

	dbPath := filepath.Join(t.TempDir(), "remove-execute.db")
	db := openCLIIntegrationSQLiteDB(t, dbPath)
	if err := menurepopkg.Migrate(db); err != nil {
		t.Fatalf("migrate menus: %v", err)
	}
	if err := casbinadapter.Migrate(db); err != nil {
		t.Fatalf("migrate casbin: %v", err)
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
	tree, err := menuSvc.Tree(context.Background())
	if err != nil {
		t.Fatalf("load menu tree: %v", err)
	}
	systemMenu := mustFindMenuByPath(t, tree, "/system")
	booksMenu, err := menuSvc.Create(context.Background(), menucommand.CreateMenu{
		ParentID:   systemMenu.ID,
		Name:       "Books",
		Path:       "/books",
		Component:  "Layout",
		Sort:       1,
		Permission: "book:view",
		Type:       "directory",
		Visible:    true,
		Enabled:    true,
		Redirect:   "/books/list",
	})
	if err != nil {
		t.Fatalf("create books menu: %v", err)
	}
	if _, err := menuSvc.Create(context.Background(), menucommand.CreateMenu{
		ParentID:   booksMenu.ID,
		Name:       "List",
		Path:       "/books/list",
		Component:  "view/book/index",
		Sort:       2,
		Permission: "book:list",
		Type:       "menu",
		Visible:    true,
		Enabled:    true,
	}); err != nil {
		t.Fatalf("create books list menu: %v", err)
	}
	store, err := casbinadapter.NewGormStore(db)
	if err != nil {
		t.Fatalf("new casbin store: %v", err)
	}
	rules := []casbinadapter.Rule{
		{Subject: "admin", Object: "/api/v1/books", Action: "GET"},
		{Subject: "admin", Object: "/api/v1/books/:id", Action: "GET"},
		{Subject: "admin", Object: "/api/v1/books", Action: "POST"},
		{Subject: "admin", Object: "/api/v1/books/:id", Action: "PUT"},
		{Subject: "admin", Object: "/api/v1/books/:id", Action: "DELETE"},
		{Subject: "admin", Object: "book", Action: "list"},
		{Subject: "admin", Object: "book", Action: "view"},
		{Subject: "admin", Object: "book", Action: "create"},
		{Subject: "admin", Object: "book", Action: "update"},
		{Subject: "admin", Object: "book", Action: "delete"},
	}
	if err := store.SavePolicies(context.Background(), rules); err != nil {
		t.Fatalf("seed casbin rules: %v", err)
	}
	policyCleanup, err := deletionapp.NewPolicyCleanupService(deletionapp.PolicyCleanupDependencies{
		ProjectRoot: root,
		BackendRoot: filepath.Join(root, "backend"),
		Store:       deletionmodel.PolicyStoreDB,
		DB:          db,
	})
	if err != nil {
		t.Fatalf("new policy cleanup service: %v", err)
	}
	output, err := captureCLIStdout(t, func() error {
		return RunWithDependencies(root, []string{"remove", "execute", "book", "--kind", "crud"}, Dependencies{
			MenuService:   menuSvc,
			PolicyCleanup: policyCleanup,
			PolicyStore:   string(deletionmodel.PolicyStoreDB),
		})
	})
	if err != nil {
		t.Fatalf("Run(remove execute) returned error: %v", err)
	}
	if !strings.Contains(output, "status=succeeded") {
		t.Fatalf("execute output missing success status:\n%s", output)
	}
	if _, err := os.Stat(filepath.Join(root, "backend", "modules", "book")); !os.IsNotExist(err) {
		t.Fatalf("expected module dir to be removed, got err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "web", "src", "views", "book")); !os.IsNotExist(err) {
		t.Fatalf("expected frontend view dir to be removed, got err=%v", err)
	}
	tree, err = menuSvc.Tree(context.Background())
	if err != nil {
		t.Fatalf("reload menu tree: %v", err)
	}
	if _, ok := mustFindMenuByPathOptional(t, tree, "/books"); ok {
		t.Fatal("expected /books menu to be removed")
	}
	if _, ok := mustFindMenuByPathOptional(t, tree, "/books/list"); ok {
		t.Fatal("expected /books/list menu to be removed")
	}
	for _, rule := range rules {
		var count int64
		if err := db.Table("casbin_rule").Where("ptype = ? AND v0 = ? AND v1 = ? AND v2 = ?", "p", rule.Subject, rule.Object, rule.Action).Count(&count).Error; err != nil {
			t.Fatalf("count policy row %v: %v", rule, err)
		}
		if count != 0 {
			t.Fatalf("expected policy row removed for %v, got %d", rule, count)
		}
	}
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
	if _, err := os.Stat(filepath.Join(root, "web", "src", "views", "system", "codegen", "index.vue")); !os.IsNotExist(err) {
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
	if !strings.Contains(output, "database preview: dry-run; no files will be written") {
		t.Fatalf("preview output missing dry-run message:\n%s", output)
	}
	if !strings.Contains(output, "planner:") {
		t.Fatalf("preview output missing planner section:\n%s", output)
	}
	if !strings.Contains(output, "resource book [crud] (books)") {
		t.Fatalf("preview output missing resource summary:\n%s", output)
	}
	if !strings.Contains(output, "field mapping") {
		t.Fatalf("preview output missing field mapping section:\n%s", output)
	}
	if !strings.Contains(output, "permission item") {
		t.Fatalf("preview output missing permission item section:\n%s", output)
	}
	if !strings.Contains(output, "file plan:") {
		t.Fatalf("preview output missing file plan section:\n%s", output)
	}
}

func TestPreviewDatabaseReportTextIncludesFieldComments(t *testing.T) {
	t.Parallel()

	report := DatabasePreviewReport{
		Source: DatabasePreviewSource{Driver: "mysql", Database: "goadmin"},
		Resources: []DatabasePreviewResource{
			{
				Name:      "order",
				Kind:      "crud",
				TableName: "orders",
				Actions:   []string{"generate"},
				Fields: []DatabasePreviewField{
					{Name: "ID", ColumnName: "id", Comment: "订单ID", SemanticType: "string", UIType: "input", Required: true},
					{Name: "Status", ColumnName: "status", Comment: "订单状态", SemanticType: "enum", UIType: "select", Required: true},
				},
			},
		},
	}

	output := previewDatabaseReportText(report)
	if !strings.Contains(output, "comment=订单ID") {
		t.Fatalf("preview output missing primary field comment:\n%s", output)
	}
	if !strings.Contains(output, "comment=订单状态") {
		t.Fatalf("preview output missing non-primary field comment:\n%s", output)
	}
	if !strings.Contains(output, "resource order [crud] (orders)") {
		t.Fatalf("preview output missing resource summary:\n%s", output)
	}
}

func TestExecuteDSLResourcesPreservesSchemaSQLFieldComments(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	resource := schema.Resource{
		Kind:   schema.KindCRUD,
		Name:   "order",
		Module: "order",
		Entity: schema.Entity{Name: "order", Fields: []schema.Field{
			{Name: "id", Type: "string", Comment: "订单ID", Primary: true},
			{Name: "tenant_id", Type: "string", Comment: "租户ID"},
			{Name: "order_no", Type: "string", Comment: "订单号"},
		}},
		Fields: []schema.Field{
			{Name: "id", Type: "string", Comment: "订单ID", Primary: true},
			{Name: "tenant_id", Type: "string", Comment: "租户ID"},
			{Name: "order_no", Type: "string", Comment: "订单号"},
		},
	}

	if err := ExecuteDSLResources(root, []schema.Resource{resource}, true); err != nil {
		t.Fatalf("ExecuteDSLResources returned error: %v", err)
	}

	schemaPath := filepath.Join(root, "backend", "modules", "order", "schema.sql")
	content, err := os.ReadFile(schemaPath)
	if err != nil {
		t.Fatalf("read schema.sql: %v", err)
	}
	text := string(content)
	if !strings.Contains(text, "`id` varchar(64) NOT NULL COMMENT '订单ID'") {
		t.Fatalf("schema.sql missing primary column comment:\n%s", text)
	}
	if !strings.Contains(text, "`tenant_id` varchar(255) NOT NULL COMMENT '租户ID'") {
		t.Fatalf("schema.sql missing non-primary column comment for tenant_id:\n%s", text)
	}
	if !strings.Contains(text, "`order_no` varchar(255) NOT NULL COMMENT '订单号'") {
		t.Fatalf("schema.sql missing non-primary column comment for order_no:\n%s", text)
	}
}

func TestRunGenerateDBGenerate(t *testing.T) {
	root := t.TempDir()
	dbPath := filepath.Join(t.TempDir(), "codegen.db")
	db := openCLIIntegrationSQLiteDB(t, dbPath)
	if err := db.Exec(`CREATE TABLE books (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL
	);`).Error; err != nil {
		t.Fatalf("create table: %v", err)
	}

	if err := Run(root, []string{
		"generate",
		"db",
		"generate",
		"--driver", "sqlite",
		"--dsn", dbPath,
		"--database", "codegen",
		"--table", "books",
	}); err != nil {
		t.Fatalf("Run(generate db generate) returned error: %v", err)
	}

	assertFileExists(t, filepath.Join(root, "backend", "modules", "book", "module.go"))
	assertFileExists(t, filepath.Join(root, "web", "src", "views", "book", "index.vue"))
	assertFileContains(t, filepath.Join(root, "backend", "modules", "book", "manifest.yaml"), "menus:")
	assertFileContains(t, filepath.Join(root, "backend", "modules", "book", "manifest.yaml"), "path: /books")
	assertFileContains(t, filepath.Join(root, "backend", "core", "auth", "casbin", "adapter", "policy.csv"), "p, admin, /api/v1/books, GET")
}

func TestRunGenerateDBValidation(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	if err := Run(root, []string{"generate", "db", "preview", "--dsn", "sqlite.db", "--database", "codegen"}); err == nil || !strings.Contains(err.Error(), "database driver is required") {
		t.Fatalf("Run(generate db preview) validation error = %v, want database driver is required", err)
	}
}

func TestExecuteDatabaseDocumentDryRunReport(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	dbPath := filepath.Join(t.TempDir(), "codegen.db")
	db := openCLIIntegrationSQLiteDB(t, dbPath)
	if err := db.Exec(`CREATE TABLE books (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL
	);`).Error; err != nil {
		t.Fatalf("create table: %v", err)
	}

	frontend := true
	policy := true
	report, err := ExecuteDatabaseDocument(root, db, nil, DatabaseExecutionRequest{
		Driver:           "sqlite",
		DSN:              dbPath,
		Database:         "codegen",
		Tables:           []string{"books"},
		GenerateFrontend: &frontend,
		GeneratePolicy:   &policy,
		MountParentPath:  "/system",
	}, true)
	if err != nil {
		t.Fatalf("ExecuteDatabaseDocument(dry-run) returned error: %v", err)
	}
	if !report.DryRun {
		t.Fatal("expected dry-run report")
	}
	if len(report.Planner.Resources) != 1 {
		t.Fatalf("expected 1 planner resource, got %d", len(report.Planner.Resources))
	}
	if got, want := report.Planner.Resources[0].Name, "book"; got != want {
		t.Fatalf("planner resource name = %q, want %q", got, want)
	}
	if len(report.Resources) != 1 {
		t.Fatalf("expected 1 preview resource, got %d", len(report.Resources))
	}
	resource := report.Resources[0]
	if len(resource.Fields) == 0 {
		t.Fatal("expected field mappings in preview report")
	}
	if len(resource.Pages) == 0 {
		t.Fatal("expected page items in preview report")
	}
	if len(resource.Permissions) == 0 {
		t.Fatal("expected permission items in preview report")
	}
	if len(report.Files) == 0 {
		t.Fatal("expected file plan entries in preview report")
	}
	if got := report.Audit.Input.MountParentPath; got != "/system" {
		t.Fatalf("audit mount parent path = %q, want %q", got, "/system")
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
	if report.Audit.Output.ConflictCount != len(report.Conflicts) {
		t.Fatalf("audit conflict count = %d, want %d", report.Audit.Output.ConflictCount, len(report.Conflicts))
	}
	if got := report.Audit.Input.Driver; got != "sqlite" {
		t.Fatalf("audit input driver = %q, want sqlite", got)
	}
	if got := report.Audit.Input.Database; got != "codegen" {
		t.Fatalf("audit input database = %q, want codegen", got)
	}
	if report.Audit.Input.DryRun != true {
		t.Fatal("expected audit dry-run input")
	}
	if secret := dbPath; strings.Contains(mustJSON(t, report), secret) {
		t.Fatalf("report leaked DSN/path %q in JSON: %s", secret, mustJSON(t, report))
	}
	if _, err := os.Stat(filepath.Join(root, "backend", "modules", "book", "module.go")); !os.IsNotExist(err) {
		t.Fatalf("dry-run should not create output files, got err=%v", err)
	}
}

func mustJSON(t *testing.T, value any) string {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal json: %v", err)
	}
	return string(data)
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

func mustFindMenuByPath(t *testing.T, items []menuModel.Menu, path string) *menuModel.Menu {
	t.Helper()
	menu, ok := mustFindMenuByPathOptional(t, items, path)
	if !ok {
		t.Fatalf("expected menu %s to exist", path)
	}
	return menu
}

func mustFindMenuByPathOptional(t *testing.T, items []menuModel.Menu, path string) (*menuModel.Menu, bool) {
	t.Helper()
	for i := range items {
		if menu, ok := findMenuNodeByPath(items[i], path); ok {
			return menu, true
		}
	}
	return nil, false
}

func findMenuNodeByPath(menu menuModel.Menu, path string) (*menuModel.Menu, bool) {
	if menu.Path == path {
		clone := menu.Clone()
		return &clone, true
	}
	for i := range menu.Children {
		if found, ok := findMenuNodeByPath(menu.Children[i], path); ok {
			return found, true
		}
	}
	return nil, false
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

func testTitleFromModule(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "Module"
	}
	parts := strings.Split(strings.ReplaceAll(value, "-", "_"), "_")
	for i, part := range parts {
		if part == "" {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, "")
}

func containsDeleteItemKind(items []deletionmodel.DeleteItem, kind deletionmodel.AssetKind) bool {
	for _, item := range items {
		if item.Kind == kind {
			return true
		}
	}
	return false
}

func createDeleteModuleFixture(t *testing.T, root, module string, includeManifest, includeUnknown bool) {
	t.Helper()
	moduleDir := filepath.Join(root, "backend", "modules", module)
	mustMkdirAll(t, filepath.Join(moduleDir, "application", "command"))
	mustMkdirAll(t, filepath.Join(moduleDir, "application", "query"))
	mustMkdirAll(t, filepath.Join(moduleDir, "application", "service"))
	mustMkdirAll(t, filepath.Join(moduleDir, "domain", "model"))
	mustMkdirAll(t, filepath.Join(moduleDir, "domain", "repository"))
	mustMkdirAll(t, filepath.Join(moduleDir, "infrastructure", "repo"))
	mustMkdirAll(t, filepath.Join(moduleDir, "transport", "http", "handler"))
	mustMkdirAll(t, filepath.Join(moduleDir, "transport", "http", "request"))
	mustMkdirAll(t, filepath.Join(moduleDir, "transport", "http", "response"))
	mustMkdirAll(t, filepath.Join(root, "backend", "core", "bootstrap"))
	mustMkdirAll(t, filepath.Join(root, "web", "src", "api"))
	mustMkdirAll(t, filepath.Join(root, "web", "src", "router", "modules"))
	mustMkdirAll(t, filepath.Join(root, "web", "src", "views", module))

	moduleTitle := testTitleFromModule(module)
	modulePlural := testPluralize(module)
	writeFixture(t, filepath.Join(moduleDir, "module.go"), "package "+module+"\n\nconst Name = \""+module+"\"\nconst ManifestPath = \"modules/"+module+"/manifest.yaml\"\n")
	writeFixture(t, filepath.Join(moduleDir, "bootstrap.go"), "// codegen:begin\npackage "+module+"\n\nfunc init() {}\n// codegen:end\n")
	if includeManifest {
		writeFixture(t, filepath.Join(moduleDir, "manifest.yaml"), strings.TrimSpace(`
# codegen:begin
name: `+module+`
version: v1
kind: crud
module: `+module+`
routes:
  - method: GET
    path: /api/v1/`+modulePlural+`
  - method: POST
    path: /api/v1/`+modulePlural+`
  - method: PUT
    path: /api/v1/`+modulePlural+`/:id
  - method: DELETE
    path: /api/v1/`+modulePlural+`/:id
  - method: GET
    path: /api/v1/`+modulePlural+`/:id
menus:
  - name: `+moduleTitle+`s
    path: /`+modulePlural+`
    component: Layout
    permission: `+module+`:view
    type: directory
    enabled: true
    visible: true
  - name: List
    path: /`+modulePlural+`/list
    parent_path: /`+modulePlural+`
    component: view/`+module+`/index
    permission: `+module+`:list
    type: menu
    enabled: true
    visible: true
permissions:
  - object: `+module+`
    action: list
  - object: `+module+`
    action: view
  - object: `+module+`
    action: create
  - object: `+module+`
    action: update
  - object: `+module+`
    action: delete
# codegen:end
`))
	}
	writeFixture(t, filepath.Join(moduleDir, "schema.sql"), "-- Database: goadmin\n")
	writeFixture(t, filepath.Join(moduleDir, "application", "command", module+".go"), "package command\n")
	writeFixture(t, filepath.Join(moduleDir, "application", "query", module+".go"), "package query\n")
	writeFixture(t, filepath.Join(moduleDir, "application", "service", "service.go"), "package service\n")
	writeFixture(t, filepath.Join(moduleDir, "domain", "model", module+".go"), "package model\n")
	writeFixture(t, filepath.Join(moduleDir, "domain", "repository", "repository.go"), "package repository\n")
	writeFixture(t, filepath.Join(moduleDir, "infrastructure", "repo", "gorm.go"), "package repo\n")
	writeFixture(t, filepath.Join(moduleDir, "transport", "http", "handler", "handler.go"), "package handler\n")
	writeFixture(t, filepath.Join(moduleDir, "transport", "http", "request", module+".go"), "package request\n")
	writeFixture(t, filepath.Join(moduleDir, "transport", "http", "response", module+".go"), "package response\n")
	if includeUnknown {
		writeFixture(t, filepath.Join(moduleDir, "notes.txt"), "manual note")
	}
	writeFixture(t, filepath.Join(root, "backend", "core", "bootstrap", "modules_gen.go"), "package bootstrap\n\nimport (\n\t\"goadmin/modules/"+module+"\"\n)\n\nfunc generatedModules() []Module {\n\treturn []Module{\n\t\t"+module+".NewBootstrap(),\n\t}\n}\n")
	writeFixture(t, filepath.Join(root, "backend", "core", "bootstrap", "modules_builtin.go"), "package bootstrap\n\nimport ()\n\nfunc builtinModules() []Module {\n\treturn []Module{}\n}\n")
	writeFixture(t, filepath.Join(root, "web", "src", "api", module+".ts"), "export {}\n")
	writeFixture(t, filepath.Join(root, "web", "src", "router", "modules", module+".ts"), "export {}\n")
	writeFixture(t, filepath.Join(root, "web", "src", "views", module, "index.vue"), "<template></template>\n")
}

func writeFixture(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func mustMkdirAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
}

func testPluralize(value string) string {
	value = strings.TrimSpace(strings.ReplaceAll(value, "-", "_"))
	if value == "" {
		return ""
	}
	switch {
	case strings.HasSuffix(value, "ch"), strings.HasSuffix(value, "sh"), strings.HasSuffix(value, "s"), strings.HasSuffix(value, "x"), strings.HasSuffix(value, "z"):
		return value + "es"
	case strings.HasSuffix(value, "y") && len(value) > 1:
		return value[:len(value)-1] + "ies"
	default:
		return value + "s"
	}
}
