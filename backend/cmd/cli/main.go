package main

import (
	"fmt"
	"os"
	"path/filepath"

	deletionapp "goadmin/codegen/application/deletion"
	codegencli "goadmin/codegen/driver/cli"
	deletionmodel "goadmin/codegen/model/deletion"
	casbinadapter "goadmin/core/auth/casbin/adapter"
	coreconfig "goadmin/core/config"
	infraDB "goadmin/infrastructure/db"
	menuservice "goadmin/modules/menu/application/service"
	menurepo "goadmin/modules/menu/infrastructure/repo"
)

func main() {
	root, err := findProjectRoot()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	runtimeDeps, depErr := loadCodegenRuntimeDependencies(root)
	if depErr != nil {
		fmt.Fprintln(os.Stderr, "warning:", depErr)
	}

	if err := codegencli.RunWithDependencies(root, os.Args[1:], runtimeDeps); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func loadCodegenRuntimeDependencies(projectRoot string) (codegencli.Dependencies, error) {
	configDir := filepath.Join(projectRoot, "backend", "config")
	if !fileExists(filepath.Join(configDir, "config.yaml")) {
		return codegencli.Dependencies{}, nil
	}
	prevConfigDir, hadConfigDir := os.LookupEnv("GOADMIN_CONFIG_DIR")
	if err := os.Setenv("GOADMIN_CONFIG_DIR", configDir); err != nil {
		return codegencli.Dependencies{}, fmt.Errorf("set GOADMIN_CONFIG_DIR: %w", err)
	}
	defer func() {
		if hadConfigDir {
			_ = os.Setenv("GOADMIN_CONFIG_DIR", prevConfigDir)
			return
		}
		_ = os.Unsetenv("GOADMIN_CONFIG_DIR")
	}()

	cfg, err := coreconfig.Load()
	if err != nil {
		return codegencli.Dependencies{}, fmt.Errorf("load config for codegen CLI: %w", err)
	}
	dbConn, err := infraDB.Open(cfg.Database)
	if err != nil {
		return codegencli.Dependencies{}, fmt.Errorf("open database for codegen CLI: %w", err)
	}
	if err := menurepo.Migrate(dbConn); err != nil {
		return codegencli.Dependencies{}, fmt.Errorf("migrate menus for codegen CLI: %w", err)
	}
	if err := casbinadapter.Migrate(dbConn); err != nil {
		return codegencli.Dependencies{}, fmt.Errorf("migrate casbin for codegen CLI: %w", err)
	}
	menuRepo, err := menurepo.NewGormRepository(dbConn)
	if err != nil {
		return codegencli.Dependencies{}, fmt.Errorf("init menu repository for codegen CLI: %w", err)
	}
	menuSvc, err := menuservice.New(menuRepo)
	if err != nil {
		return codegencli.Dependencies{}, fmt.Errorf("init menu service for codegen CLI: %w", err)
	}
	policyCleanup, err := deletionapp.NewPolicyCleanupService(deletionapp.PolicyCleanupDependencies{
		ProjectRoot: projectRoot,
		BackendRoot: filepath.Join(projectRoot, "backend"),
		Store:       deletionmodel.NormalizePolicyStoreKind(cfg.Auth.Casbin.Source),
		DB:          dbConn,
	})
	if err != nil {
		return codegencli.Dependencies{}, fmt.Errorf("init policy cleanup for codegen CLI: %w", err)
	}
	return codegencli.Dependencies{
		MenuService:   menuSvc,
		PolicyCleanup: policyCleanup,
		PolicyStore:   cfg.Auth.Casbin.Source,
	}, nil
}

func findProjectRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("detect cwd: %w", err)
	}
	current := cwd
	for {
		if fileExists(filepath.Join(current, "go.work")) {
			return current, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return cwd, nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
