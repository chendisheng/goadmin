package deletion

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	deletionmodel "goadmin/codegen/model/deletion"
)

func TestNormalizeModuleName(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"Book":                               "book",
		"backend/modules/Book/manifest.yaml": "book",
		"modules/order/module.go":            "order",
		" order ":                            "order",
		"BookProfile":                        "book_profile",
	}
	for input, want := range cases {
		if got := NormalizeModuleName(input); got != want {
			t.Fatalf("NormalizeModuleName(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestPreviewCollectsOwnedAssetsAndConflicts(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	createDeleteModuleFixture(t, root, "book", true, true)

	service := NewService(Dependencies{ProjectRoot: root, PolicyStore: "db"})
	report, err := service.Preview(deletionmodel.DeleteRequest{
		Module:       "Book",
		DryRun:       true,
		WithPolicy:   true,
		WithRuntime:  true,
		WithFrontend: true,
		WithRegistry: true,
	})
	if err != nil {
		t.Fatalf("Preview returned error: %v", err)
	}
	if report.Plan.Module != "book" {
		t.Fatalf("plan module = %q, want book", report.Plan.Module)
	}
	if report.Plan.PolicyStore != deletionmodel.PolicyStoreDB {
		t.Fatalf("policy store = %q, want db", report.Plan.PolicyStore)
	}
	if !report.Resolution.GeneratedBootstrap {
		t.Fatal("expected generated bootstrap to be detected")
	}
	if len(report.Plan.SourceFiles) == 0 {
		t.Fatal("expected source file candidates")
	}
	assertDeleteItemPath(t, report.Plan.SourceFiles, "backend/modules/book/module.go")
	assertDeleteItemPath(t, report.Plan.SourceFiles, "backend/modules/book/bootstrap.go")
	assertDeleteItemPath(t, report.Plan.SourceFiles, "backend/modules/book/manifest.yaml")
	assertDeleteItemPath(t, report.Plan.SourceFiles, "backend/modules/book/schema.sql")
	assertDeleteItemPath(t, report.Plan.RegistryChanges, "backend/core/bootstrap/modules_gen.go")
	assertDeleteItemPath(t, report.Plan.FrontendChanges, "web/src/views/book/index.vue")
	if len(report.Plan.PolicyChanges) != 10 {
		t.Fatalf("policy changes = %d, want 10", len(report.Plan.PolicyChanges))
	}
	if len(report.Plan.Conflicts) == 0 {
		t.Fatal("expected conflicts for unknown file")
	}
	if !containsConflict(report.Plan.Conflicts, "unknown-owned-file", "backend/modules/book/notes.txt") {
		t.Fatalf("expected unknown-owned-file conflict, got %#v", report.Plan.Conflicts)
	}
	if report.Plan.Summary.SourceFiles == 0 || report.Plan.Summary.RuntimeAssets == 0 || report.Plan.Summary.PolicyChanges == 0 {
		t.Fatalf("unexpected empty summary: %#v", report.Plan.Summary)
	}
}

func TestPreviewInferredManifestForGeneratedModule(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	createDeleteModuleFixture(t, root, "order", false, false)
	_ = os.Remove(filepath.Join(root, "backend", "modules", "order", "manifest.yaml"))

	service := NewService(Dependencies{ProjectRoot: root, PolicyStore: "file"})
	report, err := service.Preview(deletionmodel.DeleteRequest{Module: "order", DryRun: true, WithRuntime: true})
	if err != nil {
		t.Fatalf("Preview returned error: %v", err)
	}
	if len(report.Plan.PolicyChanges) != 0 {
		t.Fatalf("policy changes = %d, want 0 without with_policy flag", len(report.Plan.PolicyChanges))
	}
	if len(report.Plan.RuntimeAssets) == 0 {
		t.Fatal("expected runtime assets from inferred manifest")
	}
	if !containsDeleteItemKind(report.Plan.RuntimeAssets, deletionmodel.AssetKindRuntimeMenu) {
		t.Fatal("expected runtime menu candidates from inferred manifest")
	}
	if len(report.Plan.Warnings) == 0 {
		t.Fatal("expected warnings for inferred manifest preview")
	}
}

func TestDeleteRemovesGeneratedAssetsAndRefreshesRegistry(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	createDeleteModuleFixture(t, root, "book", true, false)

	service := NewService(Dependencies{ProjectRoot: root, PolicyStore: "db"})
	result, err := service.Delete(deletionmodel.DeleteRequest{
		Module:       "book",
		WithRuntime:  false,
		WithPolicy:   false,
		WithFrontend: true,
		WithRegistry: true,
	})
	if err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	if result.Status != deletionmodel.DeleteStatusSucceeded {
		t.Fatalf("status = %s, want succeeded", result.Status)
	}
	if len(result.Failures) > 0 {
		t.Fatalf("unexpected failures: %#v", result.Failures)
	}
	moduleDir := filepath.Join(root, "backend", "modules", "book")
	if _, err := os.Stat(moduleDir); !os.IsNotExist(err) {
		t.Fatalf("expected module dir to be removed, got err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "web", "src", "views", "book")); !os.IsNotExist(err) {
		t.Fatalf("expected frontend view dir to be removed, got err=%v", err)
	}
	registryPath := filepath.Join(root, "backend", "core", "bootstrap", "modules_gen.go")
	content, err := os.ReadFile(registryPath)
	if err != nil {
		t.Fatalf("read refreshed registry: %v", err)
	}
	if strings.Contains(string(content), "goadmin/modules/book") {
		t.Fatalf("registry still contains deleted module: %s", content)
	}
	if !strings.Contains(string(content), "func generatedModules() []Module") {
		t.Fatalf("registry missing generatedModules function: %s", content)
	}
}

func TestDeleteCountsRuntimeRouteCleanupWhenCoreAssetsSucceed(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	createDeleteModuleFixture(t, root, "book", true, false)

	service := NewService(Dependencies{ProjectRoot: root, PolicyStore: "db"})
	report, err := service.Preview(deletionmodel.DeleteRequest{
		Module:       "book",
		DryRun:       true,
		WithRuntime:  true,
		WithPolicy:   false,
		WithFrontend: true,
		WithRegistry: true,
	})
	if err != nil {
		t.Fatalf("Preview returned error: %v", err)
	}
	routeCount := 0
	for _, item := range report.Plan.RuntimeAssets {
		if item.Kind == deletionmodel.AssetKindRuntimeRoute {
			routeCount++
		}
	}
	if routeCount == 0 {
		t.Fatal("expected runtime route candidates in preview")
	}

	result, err := service.Delete(deletionmodel.DeleteRequest{
		Module:       "book",
		WithRuntime:  true,
		WithPolicy:   false,
		WithFrontend: true,
		WithRegistry: true,
	})
	if err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	if result.Summary.DeletedRuntimeAssets != routeCount {
		t.Fatalf("deleted runtime assets = %d, want %d", result.Summary.DeletedRuntimeAssets, routeCount)
	}
	if !containsDeleteItemKind(result.Deleted, deletionmodel.AssetKindRuntimeRoute) {
		t.Fatal("expected runtime route items to be reported as deleted")
	}
	if _, err := os.Stat(filepath.Join(root, "backend", "modules", "book")); !os.IsNotExist(err) {
		t.Fatalf("expected module dir to be removed, got err=%v", err)
	}
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

func assertDeleteItemPath(t *testing.T, items []deletionmodel.DeleteItem, want string) {
	t.Helper()
	for _, item := range items {
		if item.Path == want {
			return
		}
	}
	t.Fatalf("expected item path %q in %#v", want, items)
}

func containsConflict(conflicts []deletionmodel.DeleteConflict, kind, path string) bool {
	for _, conflict := range conflicts {
		if conflict.Kind == kind && conflict.Path == path {
			return true
		}
	}
	return false
}

func containsDeleteItemKind(items []deletionmodel.DeleteItem, kind deletionmodel.AssetKind) bool {
	for _, item := range items {
		if item.Kind == kind {
			return true
		}
	}
	return false
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
