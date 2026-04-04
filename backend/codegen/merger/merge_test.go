package merger

import (
	"strings"
	"testing"
)

func TestMergeGoContentPreservesManualDeclarations(t *testing.T) {
	t.Parallel()

	current := []byte(strings.TrimSpace(`
package demo

import (
	"fmt"
)

func Helper() string {
	return fmt.Sprint("manual")
}

type Existing struct {
	Name string
}
`))
	generated := []byte(strings.TrimSpace(`
package demo

import (
	"strings"
)

type Existing struct {
	Name string
}

func NewThing() string {
	return strings.TrimSpace(" generated ")
}
`))

	result, err := MergeContent("demo.go", current, generated, false)
	if err != nil {
		t.Fatalf("MergeContent returned error: %v", err)
	}
	if result.Conflict {
		t.Fatal("expected merged Go file without conflict")
	}
	if !result.Changed {
		t.Fatal("expected merged Go file to be marked changed")
	}
	content := string(result.Content)
	if !strings.Contains(content, "func Helper() string") {
		t.Fatalf("merged Go content lost manual helper:\n%s", content)
	}
	if !strings.Contains(content, "func NewThing() string") {
		t.Fatalf("merged Go content missing generated function:\n%s", content)
	}
	if !strings.Contains(content, `"fmt"`) || !strings.Contains(content, `"strings"`) {
		t.Fatalf("merged Go content missing merged imports:\n%s", content)
	}
}

func TestMergePolicyLinesDeduplicatesEntries(t *testing.T) {
	t.Parallel()

	current := []byte("p, admin, /api/v1/books, GET\n")
	generated := []byte("p, admin, /api/v1/books, GET\np, admin, /api/v1/books/:id, GET\n")

	result, err := MergeContent("policy.csv", current, generated, false)
	if err != nil {
		t.Fatalf("MergeContent returned error: %v", err)
	}
	content := string(result.Content)
	if strings.Count(content, "p, admin, /api/v1/books, GET") != 1 {
		t.Fatalf("expected deduplicated policy line, got:\n%s", content)
	}
	if !strings.Contains(content, "p, admin, /api/v1/books/:id, GET") {
		t.Fatalf("expected merged policy line, got:\n%s", content)
	}
}

func TestMergeYAMLContentReplacesWithIndentedGeneratedContent(t *testing.T) {
	t.Parallel()

	current := []byte(strings.TrimSpace(`
name: order
routes:
- method: GET
path: /api/v1/orders
`))
	generated := []byte(strings.TrimSpace(`
name: order
routes:
  - method: GET
    path: /api/v1/orders
  - method: POST
    path: /api/v1/orders
`))

	result, err := MergeContent("backend/modules/order/manifest.yaml", current, generated, false)
	if err != nil {
		t.Fatalf("MergeContent returned error: %v", err)
	}
	if result.Conflict {
		t.Fatal("expected YAML merge to replace content without conflict")
	}
	content := string(result.Content)
	if !strings.Contains(content, "  - method: GET") || !strings.Contains(content, "    path: /api/v1/orders") {
		t.Fatalf("expected indented YAML content, got:\n%s", content)
	}
	if strings.Contains(content, "- method: GET\npath: /api/v1/orders") {
		t.Fatalf("expected malformed YAML to be replaced, got:\n%s", content)
	}
}

func TestMergeUnsupportedTextKeepsExistingContent(t *testing.T) {
	t.Parallel()

	current := []byte("manual content\n")
	generated := []byte("generated content\n")

	result, err := MergeContent("notes.txt", current, generated, false)
	if err != nil {
		t.Fatalf("MergeContent returned error: %v", err)
	}
	if !result.Conflict {
		t.Fatal("expected unsupported merge to be reported as conflict")
	}
	if string(result.Content) != string(current) {
		t.Fatalf("unsupported merge should preserve existing content, got:\n%s", string(result.Content))
	}
}
