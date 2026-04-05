package generate

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNameHelpers(t *testing.T) {
	t.Parallel()

	t.Run("ToSnake", func(t *testing.T) {
		if got := ToSnake("UserProfile"); got != "user_profile" {
			t.Fatalf("ToSnake(UserProfile) = %q, want %q", got, "user_profile")
		}
	})

	t.Run("ToCamel", func(t *testing.T) {
		if got := ToCamel("user_profile"); got != "UserProfile" {
			t.Fatalf("ToCamel(user_profile) = %q, want %q", got, "UserProfile")
		}
	})

	t.Run("Pluralize", func(t *testing.T) {
		if got := Pluralize("category"); got != "categories" {
			t.Fatalf("Pluralize(category) = %q, want %q", got, "categories")
		}
	})
}

func TestParseFields(t *testing.T) {
	t.Parallel()

	fields, err := ParseFields("name:string,tags:[]string,publish_at:time", "", "", "")
	if err != nil {
		t.Fatalf("ParseFields returned error: %v", err)
	}
	if len(fields) != 4 {
		t.Fatalf("ParseFields len = %d, want 4", len(fields))
	}

	if fields[0].JSONName != "id" || !fields[0].Primary {
		t.Fatalf("first field = %+v, want primary id field", fields[0])
	}

	name := mustField(t, fields, "Name")
	if name.GoType != "string" {
		t.Fatalf("Name.GoType = %q, want %q", name.GoType, "string")
	}

	tags := mustField(t, fields, "Tags")
	if tags.GoType != "[]string" {
		t.Fatalf("Tags.GoType = %q, want %q", tags.GoType, "[]string")
	}

	publishAt := mustField(t, fields, "PublishAt")
	if publishAt.GoType != "time.Time" {
		t.Fatalf("PublishAt.GoType = %q, want %q", publishAt.GoType, "time.Time")
	}
}

func TestGormStringSize(t *testing.T) {
	t.Parallel()

	if got := (Field{GoType: "string", Primary: true}).GormStringSize(); got != 64 {
		t.Fatalf("primary string size = %d, want 64", got)
	}
	if got := (Field{GoType: "string", Index: true}).GormStringSize(); got != 191 {
		t.Fatalf("indexed string size = %d, want 191", got)
	}
	if got := (Field{GoType: "string"}).GormStringSize(); got != 255 {
		t.Fatalf("plain string size = %d, want 255", got)
	}
}

func TestGormTagPrimaryKeyModes(t *testing.T) {
	t.Parallel()

	stringPK := Field{GoType: "string", Primary: true, Column: "id"}.GormTag()
	if !strings.Contains(stringPK, "primaryKey") {
		t.Fatalf("string primary key tag missing primaryKey: %q", stringPK)
	}
	if !strings.Contains(stringPK, "type:varchar(64)") {
		t.Fatalf("string primary key tag missing type:varchar(64): %q", stringPK)
	}
	if !strings.Contains(stringPK, "size:64") {
		t.Fatalf("string primary key tag missing size:64: %q", stringPK)
	}
	if strings.Contains(stringPK, "autoIncrement") {
		t.Fatalf("string primary key tag should not contain autoIncrement: %q", stringPK)
	}

	intPK := Field{GoType: "int64", Primary: true, Column: "id"}.GormTag()
	if !strings.Contains(intPK, "primaryKey") {
		t.Fatalf("int primary key tag missing primaryKey: %q", intPK)
	}
	if !strings.Contains(intPK, "autoIncrement") {
		t.Fatalf("int primary key tag missing autoIncrement: %q", intPK)
	}
	if strings.Contains(intPK, "size:") {
		t.Fatalf("int primary key tag should not contain string size: %q", intPK)
	}
}

func TestGenerateModule(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	gen := New(root)
	if err := gen.GenerateModule(ModuleOptions{Name: "Inventory"}); err != nil {
		t.Fatalf("GenerateModule returned error: %v", err)
	}

	modulePath := filepath.Join(root, "backend", "modules", "inventory", "module.go")
	manifestPath := filepath.Join(root, "backend", "modules", "inventory", "manifest.yaml")
	assertFileContains(t, modulePath, "package inventory")
	assertFileContains(t, modulePath, `const Name = "inventory"`)
	assertFileContains(t, manifestPath, "kind: business-module")
	assertFileContains(t, manifestPath, "path: /api/v1/inventories")

	if _, err := os.Stat(filepath.Join(root, "backend", "modules", "inventory", "application")); !os.IsNotExist(err) {
		t.Fatalf("unexpected CRUD directory created for module scaffold: %v", err)
	}
}

func TestGenerateCRUDAndPolicyDedup(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	gen := New(root)
	fields, err := ParseFields("name:string,tags:[]string,publish_at:time", "", "", "")
	if err != nil {
		t.Fatalf("ParseFields returned error: %v", err)
	}

	opts := CRUDOptions{
		Name:             "Article",
		Fields:           fields,
		GenerateFrontend: false,
		GeneratePolicy:   true,
	}
	if err := gen.GenerateCRUD(opts); err != nil {
		t.Fatalf("GenerateCRUD returned error: %v", err)
	}
	if err := gen.GenerateCRUD(opts); err != nil {
		t.Fatalf("second GenerateCRUD returned error: %v", err)
	}

	modelPath := filepath.Join(root, "backend", "modules", "article", "domain", "model", "article.go")
	bootstrapPath := filepath.Join(root, "backend", "modules", "article", "bootstrap.go")
	requestPath := filepath.Join(root, "backend", "modules", "article", "transport", "http", "request", "article.go")
	responsePath := filepath.Join(root, "backend", "modules", "article", "transport", "http", "response", "article.go")
	policyPath := filepath.Join(root, "backend", "core", "auth", "casbin", "adapter", "policy.csv")
	registryPath := filepath.Join(root, "backend", "core", "bootstrap", "modules_gen.go")

	assertFileContains(t, modelPath, `gorm:"column:id;primaryKey;type:varchar(64);size:64"`)
	assertFileContains(t, modelPath, `gorm:"column:name;type:varchar(255);size:255"`)
	assertFileContains(t, modelPath, "PublishAt time.Time")
	assertFileContains(t, modelPath, "Tags")
	assertFileContains(t, modelPath, "[]string")
	assertFileContains(t, modelPath, "append([]string(nil), m.Tags...)")
	assertFileContains(t, modelPath, `gorm:"column:id;primaryKey;type:varchar(64);size:64"`)
	assertFileContains(t, bootstrapPath, "func NewBootstrap() corebootstrapcontract.Module")
	assertFileContains(t, bootstrapPath, "func (Bootstrap) Register(group coretransport.RouteRegistrar, deps corebootstrapcontract.Dependencies) error")

	repoPath := filepath.Join(root, "backend", "modules", "article", "infrastructure", "repo", "gorm.go")
	assertFileContains(t, repoPath, "LOWER(name) LIKE ?")
	assertFileContains(t, repoPath, "normalizePage(page, pageSize)")
	assertFileContains(t, repoPath, "Order(\"updated_at DESC, created_at DESC, id ASC\")")
	assertFileContains(t, repoPath, "strings.TrimSpace(strings.ToLower(keyword))")
	assertFileContains(t, registryPath, "article.NewBootstrap()")
	assertFileContains(t, requestPath, "Name")
	assertFileContains(t, requestPath, `json:"name,omitempty"`)
	assertFileContains(t, requestPath, `form:"name"`)
	assertFileContains(t, requestPath, "PublishAt time.Time")
	assertFileContains(t, responsePath, "type Item struct")
	assertFileContains(t, responsePath, "PublishAt time.Time")
	assertFileContains(t, policyPath, "p, admin, /api/v1/articles, GET")
	assertFileContains(t, policyPath, "p, admin, /api/v1/articles/:id, DELETE")

	content, err := os.ReadFile(policyPath)
	if err != nil {
		t.Fatalf("read policy file: %v", err)
	}
	lines := nonEmptyLines(string(content))
	if got, want := len(lines), 5; got != want {
		t.Fatalf("policy line count = %d, want %d; content=%q", got, want, string(content))
	}
}

func TestGenerateCRUDPrimaryKeyModes(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	gen := New(root)

	stringFields, err := ParseFields("name:string", "", "", "")
	if err != nil {
		t.Fatalf("ParseFields returned error: %v", err)
	}
	if err := gen.GenerateCRUD(CRUDOptions{Name: "Article", Fields: stringFields, GenerateFrontend: false, GeneratePolicy: false}); err != nil {
		t.Fatalf("GenerateCRUD(string primary) returned error: %v", err)
	}

	stringModelPath := filepath.Join(root, "backend", "modules", "article", "domain", "model", "article.go")
	stringRepoPath := filepath.Join(root, "backend", "modules", "article", "infrastructure", "repo", "gorm.go")
	assertFileContains(t, stringModelPath, `gorm:"column:id;primaryKey;type:varchar(64);size:64"`)
	assertFileContains(t, stringRepoPath, `item.Id = nextRecordID("article")`)

	numericFields, err := ParseFields("id:int64,title:string", "id", "", "")
	if err != nil {
		t.Fatalf("ParseFields returned error: %v", err)
	}
	if err := gen.GenerateCRUD(CRUDOptions{Name: "Chapter", Fields: numericFields, GenerateFrontend: false, GeneratePolicy: false}); err != nil {
		t.Fatalf("GenerateCRUD(auto increment primary) returned error: %v", err)
	}

	numericModelPath := filepath.Join(root, "backend", "modules", "chapter", "domain", "model", "chapter.go")
	numericRepoPath := filepath.Join(root, "backend", "modules", "chapter", "infrastructure", "repo", "gorm.go")
	assertFileContains(t, numericModelPath, `gorm:"column:id;primaryKey;autoIncrement"`)

	numericRepoContent, err := os.ReadFile(numericRepoPath)
	if err != nil {
		t.Fatalf("read numeric repo file: %v", err)
	}
	if strings.Contains(string(numericRepoContent), "nextRecordID(") {
		t.Fatalf("auto increment repo should not generate nextRecordID: %s", string(numericRepoContent))
	}
	if strings.Contains(string(numericRepoContent), "strings.TrimSpace(item.Id)") {
		t.Fatalf("auto increment repo should not assign string IDs: %s", string(numericRepoContent))
	}
}

func TestRefreshBootstrapRegistryFiltersGeneratedModules(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	gen := New(root)

	writeBootstrap := func(name, content string) {
		path := filepath.Join(root, "backend", "modules", name, "bootstrap.go")
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir for %s: %v", path, err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", path, err)
		}
	}

	writeBootstrap("codegen_console", strings.TrimSpace(`
package codegen_console

func NewBootstrap() any {
	return nil
}
`))
	writeBootstrap("book", strings.TrimSpace(`
// codegen:begin
package book

func NewBootstrap() any {
	return nil
}
// codegen:end
`))
	writeBootstrap("order", strings.TrimSpace(`
// codegen:begin
package order

func NewBootstrap() any {
	return nil
}
// codegen:end
`))

	if err := gen.refreshBootstrapRegistry(); err != nil {
		t.Fatalf("refreshBootstrapRegistry returned error: %v", err)
	}

	registryPath := filepath.Join(root, "backend", "core", "bootstrap", "modules_gen.go")
	assertFileContains(t, registryPath, `"goadmin/modules/book"`)
	assertFileContains(t, registryPath, `"goadmin/modules/order"`)
	content, err := os.ReadFile(registryPath)
	if err != nil {
		t.Fatalf("read registry file: %v", err)
	}
	if strings.Contains(string(content), "codegen_console") {
		t.Fatalf("registry should exclude builtin module codegen_console:\n%s", string(content))
	}
}

func TestGenerateManifestRendersMenuParentPath(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	gen := New(root)
	if err := gen.GenerateManifest(ManifestOptions{
		Name:   "Book",
		Module: "book",
		Kind:   "crud",
		Menus: []ManifestMenu{
			{Name: "Books", Path: "/books", Component: "Layout", Redirect: "/books/list", Type: "directory", Visible: true, Enabled: true, Sort: 1},
			{Name: "List", Path: "/books/list", ParentPath: "/books", Component: "view/book/index", Type: "menu", Visible: true, Enabled: true, Sort: 2},
		},
	}); err != nil {
		t.Fatalf("GenerateManifest returned error: %v", err)
	}

	manifestPath := filepath.Join(root, "backend", "modules", "book", "manifest.yaml")
	assertFileContains(t, manifestPath, "menus:")
	assertFileContains(t, manifestPath, "parent_path: /books")
	assertFileContains(t, manifestPath, "path: /books/list")
}

func TestGenerateCRUDFrontendPathsUseRepoWebRoot(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	gen := New(root)
	fields, err := ParseFields("name:string", "", "", "")
	if err != nil {
		t.Fatalf("ParseFields returned error: %v", err)
	}

	if err := gen.GenerateCRUD(CRUDOptions{Name: "Article", Fields: fields, GenerateFrontend: true, GeneratePolicy: false}); err != nil {
		t.Fatalf("GenerateCRUD returned error: %v", err)
	}

	assertPathExists(t, filepath.Join(root, "web", "src", "api", "article.ts"))
	assertPathExists(t, filepath.Join(root, "web", "src", "router", "modules", "article.ts"))
	assertPathExists(t, filepath.Join(root, "web", "src", "views", "article", "index.vue"))
	assertPathNotExists(t, filepath.Join(root, "backend", "web", "src", "api", "article.ts"))
	assertPathNotExists(t, filepath.Join(root, "backend", "web", "src", "views", "article", "index.vue"))
}

func TestGenerateCRUDFrontendRendersUsablePage(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	gen := New(root)
	fields, err := ParseFields("tenant_id:string,title:string,author:string,isbn:string,publisher:string,publish_date:time,category:string,description:string,status:string,price:int64,stock_quantity:int64,cover_image_url:string,tags:string", "", "", "")
	if err != nil {
		t.Fatalf("ParseFields returned error: %v", err)
	}

	if err := gen.GenerateCRUD(CRUDOptions{Name: "Book", Fields: fields, GenerateFrontend: true, GeneratePolicy: false}); err != nil {
		t.Fatalf("GenerateCRUD returned error: %v", err)
	}

	apiPath := filepath.Join(root, "web", "src", "api", "book.ts")
	viewPath := filepath.Join(root, "web", "src", "views", "book", "index.vue")
	assertFileContains(t, apiPath, "const basePath = '/books'")
	assertFileContains(t, apiPath, "http.post(basePath, data)")
	content, err := os.ReadFile(apiPath)
	if err != nil {
		t.Fatalf("read api file: %v", err)
	}
	if strings.Contains(string(content), "/api/v1/books") {
		t.Fatalf("generated api file still contains /api/v1 prefix: %s", string(content))
	}

	assertFileContains(t, viewPath, "AdminTable")
	assertFileContains(t, viewPath, "AdminFormDialog")
	assertFileContains(t, viewPath, "listbooks")
	assertFileContains(t, viewPath, "createBook")
	assertFileContains(t, viewPath, "updateBook")
	assertFileContains(t, viewPath, "deleteBook")
	assertFileContains(t, viewPath, "el-date-picker")
	assertFileContains(t, viewPath, "el-input-number")
	assertFileContains(t, viewPath, "prop=\"title\"")
	assertFileContains(t, viewPath, "Book管理")
}

func TestGeneratePageUsesRepoWebRoot(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	gen := New(root)
	if err := gen.GeneratePage(PageOptions{ViewScope: "system", PageName: "Report", PageSlug: "report", RoutePath: "/system/report"}); err != nil {
		t.Fatalf("GeneratePage returned error: %v", err)
	}

	assertPathExists(t, filepath.Join(root, "web", "src", "views", "system", "report.vue"))
	assertPathExists(t, filepath.Join(root, "web", "src", "router", "modules", "system-report.ts"))
	assertPathNotExists(t, filepath.Join(root, "backend", "web", "src", "views", "system", "report.vue"))
	assertPathNotExists(t, filepath.Join(root, "backend", "web", "src", "router", "modules", "system-report.ts"))
}

func TestGenerateCRUDPreservesManualGoChanges(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	gen := New(root)
	fields, err := ParseFields("name:string", "", "", "")
	if err != nil {
		t.Fatalf("ParseFields returned error: %v", err)
	}

	opts := CRUDOptions{Name: "Book", Fields: fields, GenerateFrontend: false, GeneratePolicy: false}
	if err := gen.GenerateCRUD(opts); err != nil {
		t.Fatalf("GenerateCRUD returned error: %v", err)
	}

	routerPath := filepath.Join(root, "backend", "modules", "book", "transport", "http", "router.go")
	manual := strings.TrimSpace(`

func ManualDebug() string {
	return "manual"
}
`)
	handle, err := os.OpenFile(routerPath, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatalf("open router for manual edit: %v", err)
	}
	if _, err := handle.WriteString("\n" + manual + "\n"); err != nil {
		handle.Close()
		t.Fatalf("append manual edit: %v", err)
	}
	if err := handle.Close(); err != nil {
		t.Fatalf("close router after manual edit: %v", err)
	}

	if err := gen.GenerateCRUD(opts); err != nil {
		t.Fatalf("second GenerateCRUD returned error: %v", err)
	}

	assertFileContains(t, routerPath, "func ManualDebug() string")
	assertFileContains(t, routerPath, `return "manual"`)
	assertFileContains(t, routerPath, `root.GET("", h.List)`)
	assertFileContains(t, routerPath, `import (`)
}

func TestGeneratePlugin(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	gen := New(root)
	if err := gen.GeneratePlugin(PluginOptions{Name: "demo"}); err != nil {
		t.Fatalf("GeneratePlugin returned error: %v", err)
	}

	pluginPath := filepath.Join(root, "backend", "plugin", "builtin", "demo", "demo.go")
	assertFileContains(t, pluginPath, "package demo")
	assertFileContains(t, pluginPath, "pong from demo plugin")
	assertFileContains(t, filepath.Join(root, "backend", "core", "auth", "casbin", "adapter", "policy.csv"), "p, admin, /plugins/demo/ping, GET")
}

func mustField(t *testing.T, fields []Field, name string) Field {
	t.Helper()
	for _, field := range fields {
		if field.GoName == name {
			return field
		}
	}
	t.Fatalf("field %q not found in %+v", name, fields)
	return Field{}
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

func assertPathExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected %s to exist: %v", path, err)
	}
}

func assertPathNotExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err == nil {
		t.Fatalf("expected %s to not exist", path)
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat %s: %v", path, err)
	}
}

func nonEmptyLines(content string) []string {
	var lines []string
	for _, line := range strings.Split(content, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		lines = append(lines, line)
	}
	return lines
}
