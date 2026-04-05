package bootstrap

import "testing"

func TestModulesIncludesBuiltinAndGenerated(t *testing.T) {
	t.Parallel()

	modules := Modules()
	got := make(map[string]struct{}, len(modules))
	for _, module := range modules {
		if module == nil {
			continue
		}
		got[module.Name()] = struct{}{}
	}

	for _, want := range []string{"codegen_console", "menu", "role", "user", "book", "order"} {
		if _, ok := got[want]; !ok {
			t.Fatalf("Modules() missing %q; got=%v", want, got)
		}
	}
}
