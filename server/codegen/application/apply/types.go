package apply

import (
	"context"
	"fmt"
	"strings"

	legacygenerate "goadmin/cli/generate"
	installapp "goadmin/codegen/application/install"
	lifecycle "goadmin/codegen/model/lifecycle"
)

type Logger interface {
	Printf(format string, args ...any)
}

type CacheRefresher interface {
	Reload() error
}

type Generator interface {
	GenerateModule(opts legacygenerate.ModuleOptions) error
	GenerateCRUD(opts legacygenerate.CRUDOptions) error
	GeneratePlugin(opts legacygenerate.PluginOptions) error
	GenerateManifest(opts legacygenerate.ManifestOptions) error
	AppendPolicyLines(lines []string) error
}

type Installer interface {
	InstallManifest(ctx context.Context, manifestPath string) (installapp.InstallResult, error)
}

type Dependencies struct {
	ProjectRoot string
	Generator   Generator
	Installer   Installer
	Refresher   CacheRefresher
	Logger      Logger
}

type Workload struct {
	Request      lifecycle.ApplyRequest
	Module       legacygenerate.ModuleOptions
	CRUD         legacygenerate.CRUDOptions
	Plugin       legacygenerate.PluginOptions
	Manifest     legacygenerate.ManifestOptions
	PolicyLines  []string
	ManifestPath string
}

type ExecutionError struct {
	Stage    lifecycle.ApplyFailureStage
	Category lifecycle.ApplyFailureCategory
	Err      error
}

func (e ExecutionError) Error() string {
	if e.Err == nil {
		return "apply execution failed"
	}
	stage := strings.TrimSpace(string(e.Stage))
	category := strings.TrimSpace(string(e.Category))
	if stage == "" && category == "" {
		return e.Err.Error()
	}
	if stage == "" {
		return fmt.Sprintf("%s: %v", category, e.Err)
	}
	if category == "" {
		return fmt.Sprintf("%s: %v", stage, e.Err)
	}
	return fmt.Sprintf("%s/%s: %v", category, stage, e.Err)
}

func (e ExecutionError) Unwrap() error {
	return e.Err
}

func classifyExecutionError(stage lifecycle.ApplyFailureStage, category lifecycle.ApplyFailureCategory, err error) error {
	if err == nil {
		return nil
	}
	return ExecutionError{Stage: stage, Category: category, Err: err}
}
