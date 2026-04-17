package config

import "testing"

func TestDefaultGeneratedModulesAutoMigrate(t *testing.T) {
	t.Parallel()

	cfg := Default()
	if !cfg.CodeGen.GeneratedModulesAutoMigrate {
		t.Fatal("default generated_modules_auto_migrate should be enabled")
	}

	public := cfg.Public()
	codegen, ok := public["codegen"].(map[string]any)
	if !ok {
		t.Fatalf("public codegen block missing: %#v", public["codegen"])
	}
	if got, ok := codegen["generated_modules_auto_migrate"].(bool); !ok || !got {
		t.Fatalf("public generated_modules_auto_migrate = %#v, want true", codegen["generated_modules_auto_migrate"])
	}
}
