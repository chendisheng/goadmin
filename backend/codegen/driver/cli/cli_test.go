package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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
	if !strings.Contains(report.Items[0].Actions[0], "generate page") {
		t.Fatalf("unexpected preview action: %v", report.Items[0].Actions)
	}
	if _, err := os.Stat(filepath.Join(root, "backend", "web", "src", "views", "system", "codegen", "index.vue")); !os.IsNotExist(err) {
		t.Fatalf("dry-run should not create output files, got err=%v", err)
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
