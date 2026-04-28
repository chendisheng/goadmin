package download

import (
	"archive/zip"
	codegencli "goadmin/codegen/driver/cli"
	apperrors "goadmin/core/errors"
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
  server: gin
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
		"server/modules/inventory/module.go",
		"server/modules/item/domain/model/item.go",
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

func TestServiceReturnsKeyedValidationErrors(t *testing.T) {
	t.Parallel()

	service := NewService(Dependencies{BaseDir: t.TempDir(), TTL: time.Hour})

	if _, err := service.Generate(GenerateRequest{}); err == nil {
		t.Fatal("expected validation error for empty DSL")
	} else if appErr, ok := err.(*apperrors.AppError); !ok {
		t.Fatalf("expected *AppError, got %T", err)
	} else if appErr.Key != "codegen.download.dsl_required" {
		t.Fatalf("Generate() key = %q, want codegen.download.dsl_required", appErr.Key)
	}

	if _, err := service.GenerateDatabase(nil, codegencli.DatabaseExecutionRequest{}); err == nil {
		t.Fatal("expected validation error for missing DB")
	} else if appErr, ok := err.(*apperrors.AppError); !ok {
		t.Fatalf("expected *AppError, got %T", err)
	} else if appErr.Key != "codegen.download.database_required" {
		t.Fatalf("GenerateDatabase() key = %q, want codegen.download.database_required", appErr.Key)
	}

	if _, err := service.Resolve(""); err == nil {
		t.Fatal("expected validation error for empty task id")
	} else if appErr, ok := err.(*apperrors.AppError); !ok {
		t.Fatalf("expected *AppError, got %T", err)
	} else if appErr.Key != "codegen.download.task_id_required" {
		t.Fatalf("Resolve() key = %q, want codegen.download.task_id_required", appErr.Key)
	}
}

func keys(values map[string]struct{}) []string {
	result := make([]string, 0, len(values))
	for key := range values {
		result = append(result, key)
	}
	return result
}
