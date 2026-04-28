package apply

import (
	"context"
	"testing"

	legacygenerate "goadmin/cli/generate"
	installapp "goadmin/codegen/application/install"
	lifecycle "goadmin/codegen/model/lifecycle"
)

type stubGenerator struct {
	moduleCalls   int
	crudCalls     int
	pluginCalls   int
	manifestCalls int
	policyCalls   int
}

func (g *stubGenerator) GenerateModule(opts legacygenerate.ModuleOptions) error {
	g.moduleCalls++
	return nil
}

func (g *stubGenerator) GenerateCRUD(opts legacygenerate.CRUDOptions) error {
	g.crudCalls++
	if opts.GeneratePolicy {
		g.policyCalls++
	}
	return nil
}

func (g *stubGenerator) GeneratePlugin(opts legacygenerate.PluginOptions) error {
	g.pluginCalls++
	return nil
}

func (g *stubGenerator) GenerateManifest(opts legacygenerate.ManifestOptions) error {
	g.manifestCalls++
	return nil
}

func (g *stubGenerator) AppendPolicyLines(lines []string) error {
	g.policyCalls += len(lines)
	return nil
}

type stubInstaller struct {
	manifestPath string
	calls        int
}

func (i *stubInstaller) InstallManifest(ctx context.Context, manifestPath string) (installapp.InstallResult, error) {
	i.calls++
	i.manifestPath = manifestPath
	return installapp.InstallResult{ManifestPath: manifestPath}, nil
}

type stubRefresher struct {
	calls int
}

func (r *stubRefresher) Reload() error {
	r.calls++
	return nil
}

func TestPreviewBuildsPlan(t *testing.T) {
	service := NewService(Dependencies{ProjectRoot: "/repo"})
	plan, err := service.Preview(Workload{Request: lifecycle.ApplyRequest{Module: " inventory ", Kind: " module ", PolicyStore: lifecycle.PolicyStoreCSV, Verify: true}})
	if err != nil {
		t.Fatalf("Preview returned error: %v", err)
	}
	if plan.Module != "inventory" {
		t.Fatalf("expected normalized module inventory, got %q", plan.Module)
	}
	if plan.Kind != "module" {
		t.Fatalf("expected normalized kind module, got %q", plan.Kind)
	}
	if plan.Pairing.Operation != "apply" || plan.Pairing.Counterpart != "delete" {
		t.Fatalf("unexpected pairing: %#v", plan.Pairing)
	}
	if len(plan.Validation) != 1 {
		t.Fatalf("expected one validation preview issue, got %d", len(plan.Validation))
	}
}

func TestExecuteDryRunSkipsSideEffects(t *testing.T) {
	gen := &stubGenerator{}
	inst := &stubInstaller{}
	service := NewService(Dependencies{ProjectRoot: "/repo", Generator: gen, Installer: inst})
	result, err := service.Execute(context.Background(), Workload{
		Request:  lifecycle.ApplyRequest{Module: "inventory", Kind: "module", DryRun: true, Install: true, Verify: true},
		Module:   legacygenerate.ModuleOptions{Name: "inventory"},
		Manifest: legacygenerate.ManifestOptions{Name: "inventory", Module: "inventory", Kind: "module"},
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if result.Status != lifecycle.ApplyStatusDryRun {
		t.Fatalf("expected dry run status, got %q", result.Status)
	}
	if gen.moduleCalls != 0 || gen.manifestCalls != 0 || inst.calls != 0 {
		t.Fatalf("dry run should not call generator or installer, got gen=%d manifest=%d install=%d", gen.moduleCalls, gen.manifestCalls, inst.calls)
	}
	if !result.Validation.Verified {
		t.Fatalf("dry run should mark validation as verified/skipped")
	}
}

func TestExecuteRunsGenerationInstallAndRefresh(t *testing.T) {
	gen := &stubGenerator{}
	inst := &stubInstaller{}
	refresher := &stubRefresher{}
	service := NewService(Dependencies{ProjectRoot: "/repo", Generator: gen, Installer: inst, Refresher: refresher})
	result, err := service.Execute(context.Background(), Workload{
		Request:     lifecycle.ApplyRequest{Module: "inventory", Kind: "module", Install: true, Refresh: true, PolicyStore: lifecycle.PolicyStoreCSV},
		Module:      legacygenerate.ModuleOptions{Name: "inventory"},
		Manifest:    legacygenerate.ManifestOptions{Name: "inventory", Module: "inventory", Kind: "module"},
		PolicyLines: []string{"p, admin, /api/v1/inventory, GET"},
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if result.Status != lifecycle.ApplyStatusSucceeded {
		t.Fatalf("expected succeeded status, got %q", result.Status)
	}
	if gen.moduleCalls != 1 || gen.manifestCalls != 1 || gen.policyCalls != 1 {
		t.Fatalf("unexpected generator calls: module=%d manifest=%d policy=%d", gen.moduleCalls, gen.manifestCalls, gen.policyCalls)
	}
	if inst.calls != 1 {
		t.Fatalf("expected one install call, got %d", inst.calls)
	}
	if refresher.calls != 1 {
		t.Fatalf("expected one refresh call, got %d", refresher.calls)
	}
	if result.Summary.Generated == 0 || result.Summary.Installed == 0 {
		t.Fatalf("expected summary to include generated and installed counts, got %#v", result.Summary)
	}
}

func TestExecuteCRUDUsesGeneratorOwnedPolicySync(t *testing.T) {
	gen := &stubGenerator{}
	service := NewService(Dependencies{ProjectRoot: "/repo", Generator: gen})
	result, err := service.Execute(context.Background(), Workload{
		Request: lifecycle.ApplyRequest{Module: "inventory", Kind: "crud", GeneratePolicy: true, PolicyStore: lifecycle.PolicyStoreCSV},
		CRUD: legacygenerate.CRUDOptions{
			Name:                "inventory",
			GeneratePolicy:      true,
			ManifestRoutes:      []legacygenerate.ManifestRoute{{Method: "GET", Path: "/api/v1/inventory"}},
			ManifestMenus:       []legacygenerate.ManifestMenu{{Name: "Inventory", Path: "/inventory"}},
			ManifestPermissions: []legacygenerate.ManifestPermission{{Object: "/api/v1/inventory", Action: "GET"}},
		},
		Manifest: legacygenerate.ManifestOptions{Name: "inventory", Module: "inventory", Kind: "crud"},
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if result.Status != lifecycle.ApplyStatusSucceeded {
		t.Fatalf("expected succeeded status, got %q", result.Status)
	}
	if gen.crudCalls != 1 {
		t.Fatalf("expected one CRUD generation call, got %d", gen.crudCalls)
	}
	if gen.policyCalls != 1 {
		t.Fatalf("expected exactly one generator-owned policy sync, got %d", gen.policyCalls)
	}
	if result.PolicySync.Status != "succeeded" {
		t.Fatalf("expected policy sync to be marked succeeded, got %q", result.PolicySync.Status)
	}
	if len(result.PolicySync.Synced) == 0 {
		t.Fatalf("expected policy sync result items to be populated")
	}
}

func TestPreviewPluginPolicyPathUsesSnakeCase(t *testing.T) {
	service := NewService(Dependencies{})
	plan, err := service.Preview(Workload{Request: lifecycle.ApplyRequest{Module: "plugin_admin_tools", Kind: "plugin"}, Plugin: legacygenerate.PluginOptions{Name: "PluginAdminTools"}})
	if err != nil {
		t.Fatalf("Preview returned error: %v", err)
	}
	if len(plan.PolicySync) == 0 {
		t.Fatalf("expected plugin policy sync preview items")
	}
	if got, want := plan.PolicySync[0].Path, "/plugins/plugin_admin_tools/ping"; got != want {
		t.Fatalf("expected snake-case plugin ping path %q, got %q", want, got)
	}
}

func TestExecuteFailsWhenVerificationCannotFindGeneratedManifest(t *testing.T) {
	gen := &stubGenerator{}
	service := NewService(Dependencies{ProjectRoot: t.TempDir(), Generator: gen})
	result, err := service.Execute(context.Background(), Workload{
		Request:  lifecycle.ApplyRequest{Module: "inventory", Kind: "module", Verify: true},
		Module:   legacygenerate.ModuleOptions{Name: "inventory"},
		Manifest: legacygenerate.ManifestOptions{Name: "inventory", Module: "inventory", Kind: "module"},
	})
	if err == nil {
		t.Fatalf("expected validation error, got nil")
	}
	if result.Status != lifecycle.ApplyStatusFailed {
		t.Fatalf("expected failed status, got %q", result.Status)
	}
	if len(result.Validation.Issues) == 0 {
		t.Fatalf("expected validation issues when manifest is missing")
	}
}

func TestExecuteReturnsPartialWhenVerificationAllowsPartial(t *testing.T) {
	gen := &stubGenerator{}
	service := NewService(Dependencies{ProjectRoot: t.TempDir(), Generator: gen})
	result, err := service.Execute(context.Background(), Workload{
		Request:  lifecycle.ApplyRequest{Module: "inventory", Kind: "module", Verify: true, Execution: lifecycle.ApplyExecutionRule{AllowPartial: true}},
		Module:   legacygenerate.ModuleOptions{Name: "inventory"},
		Manifest: legacygenerate.ManifestOptions{Name: "inventory", Module: "inventory", Kind: "module"},
	})
	if err != nil {
		t.Fatalf("expected partial execution to return nil error, got %v", err)
	}
	if result.Status != lifecycle.ApplyStatusPartial {
		t.Fatalf("expected partial status, got %q", result.Status)
	}
	if len(result.Validation.Issues) == 0 {
		t.Fatalf("expected validation issues when manifest is missing")
	}
}
