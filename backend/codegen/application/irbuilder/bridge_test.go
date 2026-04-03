package irbuilder

import (
	"strings"
	"testing"

	"goadmin/codegen/schema"
	dbschema "goadmin/codegen/schema/database"
)

func TestBuildSchemaDocumentFromReader(t *testing.T) {
	t.Parallel()

	stub := &stubReader{
		tables: []dbschema.Table{{
			Name:        "books",
			Schema:      "main",
			PrimaryKeys: []string{"id"},
			Columns: []dbschema.Column{
				{Name: "id", Type: "INTEGER", Primary: true, AutoIncrement: true},
				{Name: "title", Type: "TEXT", Nullable: false, Index: true},
				{Name: "author_id", Type: "INTEGER", Nullable: false},
			},
			ForeignKeys: []dbschema.ForeignKey{{
				Name:       "fk_books_author",
				Columns:    []string{"author_id"},
				RefTable:   "authors",
				RefColumns: []string{"id"},
			}},
		}},
	}
	service := NewService(Dependencies{})
	doc, err := service.BuildSchemaDocumentFromReader(stub)
	if err != nil {
		t.Fatalf("BuildSchemaDocumentFromReader returned error: %v", err)
	}
	if doc.Version != defaultIRVersion {
		t.Fatalf("expected version %q, got %q", defaultIRVersion, doc.Version)
	}
	if len(doc.Resources) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(doc.Resources))
	}
	resource := doc.Resources[0]
	if got, want := string(resource.Kind), string(schema.KindCRUD); got != want {
		t.Fatalf("resource kind = %q, want %q", got, want)
	}
	if got, want := resource.Name, "book"; got != want {
		t.Fatalf("resource name = %q, want %q", got, want)
	}
	if got, want := resource.Entity.Name, "book"; got != want {
		t.Fatalf("entity name = %q, want %q", got, want)
	}
	if !resource.GenerateFrontend || !resource.GeneratePolicy {
		t.Fatalf("expected generate frontend/policy to be enabled: %#v", resource)
	}
	if got, want := len(resource.Fields), 3; got != want {
		t.Fatalf("fields len = %d, want %d", got, want)
	}
	if got, want := len(resource.Routes), 0; got != want {
		t.Fatalf("routes len = %d, want %d", got, want)
	}
}

func TestBuildDSLAndPlanFromReader(t *testing.T) {
	t.Parallel()

	stub := &stubReader{
		tables: []dbschema.Table{{
			Name:    "books",
			Schema:  "main",
			Columns: []dbschema.Column{{Name: "id", Type: "INTEGER", Primary: true, AutoIncrement: true}, {Name: "title", Type: "TEXT", Nullable: false}},
		}},
	}
	service := NewService(Dependencies{})
	dslBytes, err := service.BuildDSLDocumentFromReader(stub)
	if err != nil {
		t.Fatalf("BuildDSLDocumentFromReader returned error: %v", err)
	}
	if !strings.Contains(string(dslBytes), "resources:") {
		t.Fatalf("expected DSL output to contain resources block, got:\n%s", string(dslBytes))
	}
	parsed, err := schema.ParseYAML(dslBytes)
	if err != nil {
		t.Fatalf("ParseYAML(dslBytes) returned error: %v", err)
	}
	plan, err := service.PlanSchemaDocument(parsed)
	if err != nil {
		t.Fatalf("PlanSchemaDocument returned error: %v", err)
	}
	if len(plan.Resources) != 1 {
		t.Fatalf("expected 1 planned resource, got %d", len(plan.Resources))
	}
	if got, want := plan.Resources[0].Name, "book"; got != want {
		t.Fatalf("planned resource name = %q, want %q", got, want)
	}
}
