package deleteapp

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	lifecycle "goadmin/codegen/model/lifecycle"
)

func TestNormalizeModuleName(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"Book":                              "book",
		"server/modules/Book/manifest.yaml": "book",
		"modules/order/module.go":           "order",
		" order ":                           "order",
		"BookProfile":                       "book_profile",
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
	report, err := service.Preview(lifecycle.DeleteRequest{
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
	if report.Plan.PolicyStore != lifecycle.PolicyStoreDB {
		t.Fatalf("policy store = %q, want db", report.Plan.PolicyStore)
	}
	if !report.Resolution.GeneratedBootstrap {
		t.Fatal("expected generated bootstrap to be detected")
	}
	if len(report.Plan.SourceFiles) == 0 {
		t.Fatal("expected source file candidates")
	}
	assertDeleteItemPath(t, report.Plan.SourceFiles, "server/modules/book/module.go")
	assertDeleteItemPath(t, report.Plan.SourceFiles, "server/modules/book/bootstrap.go")
	assertDeleteItemPath(t, report.Plan.SourceFiles, "server/modules/book/manifest.yaml")
	assertDeleteItemPath(t, report.Plan.SourceFiles, "server/modules/book/schema.sql")
	assertDeleteItemPath(t, report.Plan.SourceFiles, "server/modules/book/locales/zh-CN/book.yaml")
	assertDeleteItemPath(t, report.Plan.SourceFiles, "server/modules/book/locales/en-US/book.yaml")
	assertDeleteItemPath(t, report.Plan.RegistryChanges, "server/core/bootstrap/modules_gen.go")
	assertDeleteItemPath(t, report.Plan.FrontendChanges, "web/src/views/book/index.vue")
	assertDeleteItemPath(t, report.Plan.FrontendChanges, "web/src/views/book/index.vue.js")
	assertDeleteItemPath(t, report.Plan.FrontendChanges, "web/src/api/book.js")
	assertDeleteItemPath(t, report.Plan.FrontendChanges, "web/src/router/modules/book.js")
	if len(report.Plan.PolicyChanges) != 10 {
		t.Fatalf("policy changes = %d, want 10", len(report.Plan.PolicyChanges))
	}
	if len(report.Plan.Conflicts) == 0 {
		t.Fatal("expected conflicts for unknown file")
	}
	if !containsConflict(report.Plan.Conflicts, "unknown-owned-file", "server/modules/book/notes.txt") {
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
	_ = os.Remove(filepath.Join(root, "server", "modules", "order", "manifest.yaml"))

	service := NewService(Dependencies{ProjectRoot: root, PolicyStore: "file"})
	report, err := service.Preview(lifecycle.DeleteRequest{Module: "order", DryRun: true, WithRuntime: true})
	if err != nil {
		t.Fatalf("Preview returned error: %v", err)
	}
	if len(report.Plan.PolicyChanges) != 0 {
		t.Fatalf("policy changes = %d, want 0 without with_policy flag", len(report.Plan.PolicyChanges))
	}
	if len(report.Plan.RuntimeAssets) == 0 {
		t.Fatal("expected runtime assets from inferred manifest")
	}
	if !containsDeleteItemKind(report.Plan.RuntimeAssets, lifecycle.AssetKindRuntimeMenu) {
		t.Fatal("expected runtime menu candidates from inferred manifest")
	}
	if len(report.Plan.Warnings) == 0 {
		t.Fatal("expected warnings for inferred manifest preview")
	}
}

func TestPlanReturnsPreviewPlan(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	createDeleteModuleFixture(t, root, "book", true, false)

	service := NewService(Dependencies{ProjectRoot: root, PolicyStore: "db"})
	plan, err := service.Plan(lifecycle.DeleteRequest{
		Module:       "Book",
		WithPolicy:   true,
		WithFrontend: true,
		WithRegistry: true,
	})
	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}
	if plan.Module != "book" {
		t.Fatalf("plan module = %q, want book", plan.Module)
	}
	if plan.PolicyStore != lifecycle.PolicyStoreDB {
		t.Fatalf("policy store = %q, want db", plan.PolicyStore)
	}
	if len(plan.SourceFiles) == 0 || len(plan.PolicyChanges) == 0 || len(plan.RegistryChanges) == 0 {
		t.Fatalf("expected populated delete plan, got %#v", plan)
	}
	if plan.Summary.Total == 0 {
		t.Fatalf("expected non-zero plan summary, got %#v", plan.Summary)
	}
}

func TestDeleteRemovesGeneratedAssetsAndRefreshesRegistry(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	createDeleteModuleFixture(t, root, "book", true, false)

	service := NewService(Dependencies{ProjectRoot: root, PolicyStore: "db"})
	result, err := service.Delete(lifecycle.DeleteRequest{
		Module:       "book",
		WithRuntime:  false,
		WithPolicy:   false,
		WithFrontend: true,
		WithRegistry: true,
	})
	if err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	if result.Status != lifecycle.DeleteStatusSucceeded {
		t.Fatalf("status = %s, want succeeded", result.Status)
	}
	if len(result.Failures) > 0 {
		t.Fatalf("unexpected failures: %#v", result.Failures)
	}
	if result.Audit.Operation != "delete" || result.Audit.Module != "book" {
		t.Fatalf("unexpected audit metadata: %#v", result.Audit)
	}
	if result.Audit.Failures.Total != 0 {
		t.Fatalf("expected zero audit failures, got %#v", result.Audit.Failures)
	}
	if !result.Validation.Verified || result.Validation.Status != "passed" {
		t.Fatalf("expected validation to pass, got %#v", result.Validation)
	}
	moduleDir := filepath.Join(root, "server", "modules", "book")
	if _, err := os.Stat(moduleDir); !os.IsNotExist(err) {
		t.Fatalf("expected module dir to be removed, got err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "web", "src", "views", "book")); !os.IsNotExist(err) {
		t.Fatalf("expected frontend view dir to be removed, got err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "web", "src", "views", "book", "index.vue.js")); !os.IsNotExist(err) {
		t.Fatalf("expected vue companion js file to be removed, got err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "web", "src", "api", "book.js")); !os.IsNotExist(err) {
		t.Fatalf("expected api companion js file to be removed, got err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "web", "src", "router", "modules", "book.js")); !os.IsNotExist(err) {
		t.Fatalf("expected router companion js file to be removed, got err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "server", "modules", "book", "locales", "zh-CN", "book.yaml")); !os.IsNotExist(err) {
		t.Fatalf("expected zh-CN locale file to be removed, got err=%v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "server", "modules", "book", "locales", "en-US", "book.yaml")); !os.IsNotExist(err) {
		t.Fatalf("expected en-US locale file to be removed, got err=%v", err)
	}
	registryPath := filepath.Join(root, "server", "core", "bootstrap", "modules_gen.go")
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

func TestDeleteBlockedByConflictsProducesValidationAudit(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	createDeleteModuleFixture(t, root, "book", true, true)

	service := NewService(Dependencies{ProjectRoot: root, PolicyStore: "db"})
	result, err := service.Delete(lifecycle.DeleteRequest{
		Module:       "book",
		WithRuntime:  true,
		WithPolicy:   true,
		WithFrontend: true,
		WithRegistry: true,
	})
	if err == nil {
		t.Fatal("expected delete to fail because fixture contains an unknown owned file")
	}
	if result.Status != lifecycle.DeleteStatusFailed {
		t.Fatalf("status = %s, want failed", result.Status)
	}
	if len(result.Failures) == 0 {
		t.Fatal("expected validation failure to be recorded")
	}
	if result.Failures[0].Category != lifecycle.DeleteFailureCategoryValidation {
		t.Fatalf("failure category = %s, want validation", result.Failures[0].Category)
	}
	if result.Audit.Failures.Validation == 0 {
		t.Fatalf("expected validation failures to be counted in audit: %#v", result.Audit.Failures)
	}
	if result.Validation.Verified {
		t.Fatalf("expected validation to fail, got %#v", result.Validation)
	}
}

func TestValidateDeleteExecutionDetectsResidualFile(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	createDeleteModuleFixture(t, root, "book", true, false)

	service := NewService(Dependencies{ProjectRoot: root, PolicyStore: "db"})
	report := service.validateDeleteExecution(context.Background(), lifecycle.DeletePlan{
		Module:      "book",
		PolicyStore: lifecycle.PolicyStoreDB,
	}, []lifecycle.DeleteItem{{
		Module: "book",
		Kind:   lifecycle.AssetKindSourceFile,
		Path:   "server/modules/book/module.go",
	}}, nil)
	if report.Verified {
		t.Fatalf("expected validation to fail for residual file, got %#v", report)
	}
	if report.Status != "failed" {
		t.Fatalf("status = %s, want failed", report.Status)
	}
	if len(report.Issues) == 0 {
		t.Fatal("expected validation issues for residual file")
	}
	if report.Issues[0].Category != lifecycle.DeleteFailureCategoryFile {
		t.Fatalf("issue category = %s, want file", report.Issues[0].Category)
	}
	if !strings.Contains(report.Issues[0].Message, "still exists") {
		t.Fatalf("expected residual file message, got %#v", report.Issues[0])
	}
}

func TestDeleteCountsRuntimeRouteCleanupWhenCoreAssetsSucceed(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	createDeleteModuleFixture(t, root, "book", true, false)

	service := NewService(Dependencies{ProjectRoot: root, PolicyStore: "db"})
	report, err := service.Preview(lifecycle.DeleteRequest{
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
		if item.Kind == lifecycle.AssetKindRuntimeRoute {
			routeCount++
		}
	}
	if routeCount == 0 {
		t.Fatal("expected runtime route candidates in preview")
	}

	result, err := service.Delete(lifecycle.DeleteRequest{
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
	if !containsDeleteItemKind(result.Deleted, lifecycle.AssetKindRuntimeRoute) {
		t.Fatal("expected runtime route items to be reported as deleted")
	}
	if _, err := os.Stat(filepath.Join(root, "server", "modules", "book")); !os.IsNotExist(err) {
		t.Fatalf("expected module dir to be removed, got err=%v", err)
	}
}

func TestDeleteReportsPolicyCleanupExecutionFailure(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	createDeleteModuleFixture(t, root, "book", true, false)

	service := NewService(Dependencies{ProjectRoot: root, PolicyStore: "db"})
	service.policyCleanup = &PolicyCleanupService{store: &failingPolicyStore{
		kind:      lifecycle.PolicyStoreDB,
		deleteErr: errors.New("policy cleanup failed"),
	}}

	result, err := service.Delete(lifecycle.DeleteRequest{
		Module:       "book",
		WithPolicy:   true,
		WithFrontend: true,
		WithRegistry: true,
	})
	if err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	if result.Status != lifecycle.DeleteStatusPartial {
		t.Fatalf("status = %s, want partial", result.Status)
	}
	if !containsFailure(result.Failures, lifecycle.DeleteFailureCategoryDatabase, lifecycle.DeleteFailureStagePolicy) {
		t.Fatalf("expected policy execution failure, got %#v", result.Failures)
	}
	if !result.Validation.Verified {
		t.Fatalf("expected validation to remain verified after policy cleanup error, got %#v", result.Validation)
	}
	if result.Audit.Failures.Database == 0 {
		t.Fatalf("expected database failure count in audit, got %#v", result.Audit.Failures)
	}
}

func createDeleteModuleFixture(t *testing.T, root, module string, includeManifest, includeUnknown bool) {
	t.Helper()
	moduleDir := filepath.Join(root, "server", "modules", module)
	mustMkdirAll(t, filepath.Join(moduleDir, "application", "command"))
	mustMkdirAll(t, filepath.Join(moduleDir, "application", "query"))
	mustMkdirAll(t, filepath.Join(moduleDir, "application", "service"))
	mustMkdirAll(t, filepath.Join(moduleDir, "domain", "model"))
	mustMkdirAll(t, filepath.Join(moduleDir, "domain", "repository"))
	mustMkdirAll(t, filepath.Join(moduleDir, "infrastructure", "repo"))
	mustMkdirAll(t, filepath.Join(moduleDir, "transport", "http", "handler"))
	mustMkdirAll(t, filepath.Join(moduleDir, "transport", "http", "request"))
	mustMkdirAll(t, filepath.Join(moduleDir, "transport", "http", "response"))
	mustMkdirAll(t, filepath.Join(moduleDir, "locales", "zh-CN"))
	mustMkdirAll(t, filepath.Join(moduleDir, "locales", "en-US"))
	mustMkdirAll(t, filepath.Join(root, "server", "core", "bootstrap"))
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
	writeFixture(t, filepath.Join(moduleDir, "locales", "zh-CN", module+".yaml"), "title: zh-CN\n")
	writeFixture(t, filepath.Join(moduleDir, "locales", "en-US", module+".yaml"), "title: en-US\n")
	if includeUnknown {
		writeFixture(t, filepath.Join(moduleDir, "notes.txt"), "manual note")
	}
	writeFixture(t, filepath.Join(root, "server", "core", "bootstrap", "modules_gen.go"), "package bootstrap\n\nimport (\n\t\"goadmin/modules/"+module+"\"\n)\n\nfunc generatedModules() []Module {\n\treturn []Module{\n\t\t"+module+".NewBootstrap(),\n\t}\n}\n")
	writeFixture(t, filepath.Join(root, "server", "core", "bootstrap", "modules_builtin.go"), "package bootstrap\n\nimport ()\n\nfunc builtinModules() []Module {\n\treturn []Module{}\n}\n")
	writeFixture(t, filepath.Join(root, "web", "src", "api", module+".ts"), "export {}\n")
	writeFixture(t, filepath.Join(root, "web", "src", "api", module+".js"), "export {}\n")
	writeFixture(t, filepath.Join(root, "web", "src", "router", "modules", module+".ts"), "export {}\n")
	writeFixture(t, filepath.Join(root, "web", "src", "router", "modules", module+".js"), "export {}\n")
	writeFixture(t, filepath.Join(root, "web", "src", "views", module, "index.vue"), "<template></template>\n")
	writeFixture(t, filepath.Join(root, "web", "src", "views", module, "index.vue.js"), "export {}\n")
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

func assertDeleteItemPath(t *testing.T, items []lifecycle.DeleteItem, want string) {
	t.Helper()
	for _, item := range items {
		if item.Path == want {
			return
		}
	}
	t.Fatalf("expected item path %q in %#v", want, items)
}

func containsConflict(conflicts []lifecycle.DeleteConflict, kind, path string) bool {
	for _, conflict := range conflicts {
		if conflict.Kind == kind && conflict.Path == path {
			return true
		}
	}
	return false
}

func containsDeleteItemKind(items []lifecycle.DeleteItem, kind lifecycle.AssetKind) bool {
	for _, item := range items {
		if item.Kind == kind {
			return true
		}
	}
	return false
}

func containsFailure(failures []lifecycle.DeleteFailure, category lifecycle.DeleteFailureCategory, stage lifecycle.DeleteFailureStage) bool {
	for _, failure := range failures {
		if failure.Category == category && failure.Stage == stage {
			return true
		}
	}
	return false
}

type failingPolicyStore struct {
	kind      lifecycle.PolicyStoreKind
	deleteErr error
}

func (s *failingPolicyStore) Kind() lifecycle.PolicyStoreKind { return s.kind }

func (s *failingPolicyStore) ListByModule(context.Context, string) ([]lifecycle.PolicyAsset, error) {
	return []lifecycle.PolicyAsset{{
		Store:     s.kind,
		Module:    "book",
		SourceRef: "casbin_rule:1",
		PType:     "p",
		V0:        "book",
		V1:        "book",
		V2:        "read",
		Managed:   true,
	}}, nil
}

func (s *failingPolicyStore) DeleteBySelector(context.Context, lifecycle.PolicySelector) (int, error) {
	return 0, s.deleteErr
}

func (s *failingPolicyStore) Validate(context.Context) error { return nil }

func (s *failingPolicyStore) Preview(ctx context.Context, req PolicyCleanupRequest) (PolicyCleanupPreview, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	_ = ctx
	if len(req.Selectors) == 0 {
		req.Selectors = []lifecycle.PolicySelector{{Module: req.Module, Store: s.kind, Metadata: map[string]any{"managed": true}, V0: req.Module, V1: req.Module, V2: "read"}}
	}
	selector := req.Selectors[0]
	item := PolicyCleanupItem{
		Selector:   selector,
		Rule:       lifecycle.PolicyAsset{Store: s.kind, Module: req.Module, SourceRef: "casbin_rule:1", PType: "p", V0: selector.V0, V1: selector.V1, V2: selector.V2, Managed: true},
		Decision:   "delete",
		Reason:     "matched by structured selector",
		MatchCount: 1,
	}
	return PolicyCleanupPreview{
		Request: req,
		Store:   s.kind,
		Items:   []PolicyCleanupItem{item},
		Summary: PolicyCleanupSummary{Server: string(s.kind), Total: 1, Selected: 1, Deleted: 1, Validated: true},
	}, nil
}

func (s *failingPolicyStore) Delete(context.Context, PolicyCleanupRequest) (PolicyCleanupResult, error) {
	return PolicyCleanupResult{}, s.deleteErr
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
