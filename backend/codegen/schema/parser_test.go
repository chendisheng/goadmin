package schema

import (
	"strings"
	"testing"
)

func TestParseYAMLDocument(t *testing.T) {
	t.Parallel()

	input := []byte(strings.TrimSpace(`
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
    - name: status
      type: string
      ui_type: radio
      enum:
        values:
          - draft
          - published
pages:
  - list
  - form
permissions:
  - inventory:view
  - inventory:edit
`))

	doc, err := ParseYAML(input)
	if err != nil {
		t.Fatalf("ParseYAML returned error: %v", err)
	}

	if got, want := doc.Module, "inventory"; got != want {
		t.Fatalf("Module = %q, want %q", got, want)
	}
	if got, want := string(doc.Kind), string(KindBusinessModule); got != want {
		t.Fatalf("Kind = %q, want %q", got, want)
	}
	if got, want := doc.Framework.Backend, "gin"; got != want {
		t.Fatalf("Framework.Backend = %q, want %q", got, want)
	}
	if got, want := doc.Framework.Frontend, "vue3"; got != want {
		t.Fatalf("Framework.Frontend = %q, want %q", got, want)
	}
	if got, want := doc.Entity.Name, "item"; got != want {
		t.Fatalf("Entity.Name = %q, want %q", got, want)
	}
	if got, want := len(doc.Entity.Fields), 3; got != want {
		t.Fatalf("Entity.Fields len = %d, want %d", got, want)
	}
	if got, want := doc.Entity.Fields[2].UIType, "radio"; got != want {
		t.Fatalf("Entity.Fields[2].UIType = %q, want %q", got, want)
	}
	if doc.Entity.Fields[2].Enum == nil || len(doc.Entity.Fields[2].Enum.Values) != 2 {
		t.Fatalf("Entity.Fields[2].Enum = %#v, want 2 values", doc.Entity.Fields[2].Enum)
	}
	if got, want := len(doc.Pages), 2; got != want {
		t.Fatalf("Pages len = %d, want %d", got, want)
	}
	if got, want := len(doc.Permissions), 2; got != want {
		t.Fatalf("Permissions len = %d, want %d", got, want)
	}
	if got, want := len(doc.Resources), 1; got != want {
		t.Fatalf("Resources len = %d, want %d", got, want)
	}
	resource := doc.Resources[0]
	if got, want := string(resource.Kind), string(KindBusinessModule); got != want {
		t.Fatalf("Resource.Kind = %q, want %q", got, want)
	}
	if got, want := resource.Name, "item"; got != want {
		t.Fatalf("Resource.Name = %q, want %q", got, want)
	}
	if !resource.GenerateFrontend {
		t.Fatalf("Resource.GenerateFrontend = false, want true")
	}
}

func TestParseYAMLRejectsUnknownField(t *testing.T) {
	t.Parallel()

	input := []byte(strings.TrimSpace(`
module: inventory
kind: business-module
unknown: value
`))

	if _, err := ParseYAML(input); err == nil {
		t.Fatal("ParseYAML succeeded, want error for unknown field")
	}
}
