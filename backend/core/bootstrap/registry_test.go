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

	for _, want := range []string{"casbin", "codegen_console", "menu", "role", "upload", "user", "book"} {
		if _, ok := got[want]; !ok {
			t.Fatalf("Modules() missing %q; got=%v", want, got)
		}
	}
}

func TestBuiltinAndGeneratedModulesSplit(t *testing.T) {
	t.Parallel()

	builtin := BuiltinModules()
	generated := GeneratedModules()
	if len(builtin) == 0 {
		t.Fatal("BuiltinModules returned no modules")
	}
	if len(generated) == 0 {
		t.Fatal("GeneratedModules returned no modules")
	}

	builtinNames := make(map[string]struct{}, len(builtin))
	for _, module := range builtin {
		if module == nil {
			continue
		}
		builtinNames[module.Name()] = struct{}{}
	}
	for _, want := range []string{"codegen_console", "dictionary", "menu", "role", "upload", "user"} {
		if _, ok := builtinNames[want]; !ok {
			t.Fatalf("BuiltinModules() missing %q; got=%v", want, builtinNames)
		}
	}

	generatedNames := make(map[string]struct{}, len(generated))
	for _, module := range generated {
		if module == nil {
			continue
		}
		generatedNames[module.Name()] = struct{}{}
	}
	for _, want := range []string{"book"} {
		if _, ok := generatedNames[want]; !ok {
			t.Fatalf("GeneratedModules() missing %q; got=%v", want, generatedNames)
		}
	}
}
