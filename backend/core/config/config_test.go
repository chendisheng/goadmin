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

func TestDefaultDatabaseLogSQL(t *testing.T) {
	t.Parallel()

	cfg := Default()
	if cfg.Database.LogSQL {
		t.Fatal("default database log_sql should be disabled")
	}

	public := cfg.Public()
	database, ok := public["database"].(map[string]any)
	if !ok {
		t.Fatalf("public database block missing: %#v", public["database"])
	}
	if got, ok := database["log_sql"].(bool); !ok || got {
		t.Fatalf("public log_sql = %#v, want false", database["log_sql"])
	}
}
