package download

import (
	"archive/zip"
	"os"
	"strings"
	"testing"
	"time"
)

func TestServiceGenerateAndResolve(t *testing.T) {
	t.Parallel()

	service := NewService(Dependencies{
		BaseDir: t.TempDir(),
		TTL:     time.Hour,
	})
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
permissions:
  - inventory:view
`)

	artifact, err := service.Generate(GenerateRequest{
		DSL:           dsl,
		PackageName:   "inventory-module",
		IncludeReadme: true,
		IncludeReport: true,
		IncludeDSL:    true,
	})
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	if artifact.TaskID == "" {
		t.Fatal("expected task id")
	}
	if artifact.DownloadURL == "" {
		t.Fatal("expected download url")
	}
	if artifact.SizeBytes <= 0 {
		t.Fatalf("expected positive package size, got %d", artifact.SizeBytes)
	}
	if artifact.FileCount <= 0 {
		t.Fatalf("expected positive file count, got %d", artifact.FileCount)
	}

	resolved, err := service.Resolve(artifact.TaskID)
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}
	if resolved.PackagePath == "" {
		t.Fatal("expected package path")
	}
	if resolved.Filename == "" {
		t.Fatal("expected filename")
	}

	archive, err := zip.OpenReader(resolved.PackagePath)
	if err != nil {
		t.Fatalf("OpenReader returned error: %v", err)
	}
	defer archive.Close()

	entries := make(map[string]struct{}, len(archive.File))
	for _, file := range archive.File {
		entries[file.Name] = struct{}{}
	}
	for _, want := range []string{
		"README.md",
		"dsl.yaml",
		"generation-report.json",
		"changes.txt",
		"backend/modules/inventory/module.go",
		"backend/modules/item/domain/model/item.go",
		"web/src/views/item/index.vue",
	} {
		if _, ok := entries[want]; !ok {
			t.Fatalf("zip missing %s; entries=%v", want, keys(entries))
		}
	}

	if _, err := os.Stat(resolved.PackagePath); err != nil {
		t.Fatalf("stat package path: %v", err)
	}
}

func keys(values map[string]struct{}) []string {
	result := make([]string, 0, len(values))
	for key := range values {
		result = append(result, key)
	}
	return result
}
