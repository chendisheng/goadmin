package postprocess

import (
	"strings"
	"testing"
)

func TestPatchGoRouterRegistration(t *testing.T) {
	t.Parallel()

	source := []byte(strings.TrimSpace(`
package router

import (
	"fmt"
)

func Register(group any) {
	fmt.Println("hello")
}
`))

	patched, err := PatchGoRouterRegistration(source, "bookhttp", "goadmin/modules/book/transport/http", `bookhttp.Register(protected, bookhttp.Dependencies{})`)
	if err != nil {
		t.Fatalf("PatchGoRouterRegistration returned error: %v", err)
	}
	content := string(patched)
	if !strings.Contains(content, `"goadmin/modules/book/transport/http"`) {
		t.Fatalf("patched router missing import:\n%s", content)
	}
	if !strings.Contains(normalizeStatement(content), normalizeStatement(`bookhttp.Register(protected, bookhttp.Dependencies{})`)) {
		t.Fatalf("patched router missing registration call:\n%s", content)
	}
	if !strings.Contains(content, `fmt.Println("hello")`) {
		t.Fatalf("patched router lost existing body:\n%s", content)
	}
}
