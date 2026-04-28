package lifecycle

import (
	"encoding/json"
	"testing"
	"time"
)

func TestDeleteRequestNormalizeAndValidate(t *testing.T) {
	req := DeleteRequest{
		Module:        "  inventory  ",
		Kind:          "  business-module  ",
		PolicyStore:   NormalizePolicyStoreKind(" CSV "),
		Compatibility: LegacyCompatibilityRule{Mode: NormalizeLegacyCompatibilityMode(" template-infer ")},
	}

	normalized := req.Normalize()
	if normalized.Module != "inventory" {
		t.Fatalf("expected normalized module inventory, got %q", normalized.Module)
	}
	if normalized.Kind != "business-module" {
		t.Fatalf("expected normalized kind business-module, got %q", normalized.Kind)
	}
	if normalized.PolicyStore != PolicyStoreCSV {
		t.Fatalf("expected csv policy store, got %q", normalized.PolicyStore)
	}
	if normalized.Compatibility.Mode != LegacyCompatibilityModeTemplateInfer {
		t.Fatalf("expected template_infer compatibility mode, got %q", normalized.Compatibility.Mode)
	}
	if err := normalized.Validate(); err != nil {
		t.Fatalf("expected request to validate, got error: %v", err)
	}
}

func TestModuleOwnershipAndPlanRoundTrip(t *testing.T) {
	ownership := ModuleOwnership{
		Module:           "inventory",
		Kind:             "business-module",
		GeneratorVersion: "v1.0.0",
		GeneratedAt:      time.Unix(1700000000, 0).UTC(),
		ManifestPath:     "server/modules/inventory/codegen.manifest.json",
		ManifestFormat:   "json",
		Source:           "dsl",
		OwnedFiles: []DeleteItem{{
			Module:  "inventory",
			Kind:    AssetKindSourceFile,
			Path:    "server/modules/inventory/module.go",
			Origin:  AssetOriginGenerated,
			Managed: true,
		}},
		RuntimeAssets: []DeleteItem{{
			Module:  "inventory",
			Kind:    AssetKindRuntimeMenu,
			Ref:     "system:inventory",
			Origin:  AssetOriginGenerated,
			Managed: true,
		}},
		PolicyAssets: []PolicyAsset{
			{
				Store:       PolicyStoreCSV,
				Module:      "inventory",
				SourceRef:   "inventory:view",
				PType:       "p",
				V0:          "inventory",
				V1:          "/api/v1/inventory",
				V2:          "read",
				Managed:     true,
				GeneratedAt: time.Unix(1700000000, 0).UTC(),
			},
			{
				Store:     PolicyStoreDB,
				Module:    "inventory",
				SourceRef: "inventory:edit",
				PType:     "p",
				V0:        "inventory",
				V1:        "/api/v1/inventory",
				V2:        "update",
				Managed:   true,
			},
		},
		FrontendAssets: []DeleteItem{{
			Module:  "inventory",
			Kind:    AssetKindFrontendFile,
			Path:    "web/src/views/inventory/index.vue",
			Origin:  AssetOriginGenerated,
			Managed: true,
		}},
		Compatibility: LegacyCompatibilityRule{
			Mode:                   LegacyCompatibilityModeConservative,
			RequireManifest:        true,
			RequireExplicitConfirm: true,
			AllowPathInference:     true,
			ManifestPaths:          []string{"server/modules/inventory/codegen.manifest.json"},
			ModuleRoots:            []string{"server/modules/inventory"},
			OwnedFilePatterns:      []string{"server/modules/inventory/**"},
			FallbackPolicyStores:   []PolicyStoreKind{PolicyStoreCSV, PolicyStoreDB},
			Notes:                  []string{"legacy module fallback"},
		},
		Metadata: map[string]any{"owner": "codegen"},
	}

	payload, err := json.Marshal(ownership)
	if err != nil {
		t.Fatalf("marshal ownership: %v", err)
	}
	var decoded ModuleOwnership
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("unmarshal ownership: %v", err)
	}
	if len(decoded.PolicyAssets) != 2 {
		t.Fatalf("expected 2 policy assets, got %d", len(decoded.PolicyAssets))
	}
	if decoded.PolicyAssets[0].Selector().Store != PolicyStoreCSV {
		t.Fatalf("expected csv selector store, got %q", decoded.PolicyAssets[0].Selector().Store)
	}
	if decoded.PolicyAssets[1].Selector().Store != PolicyStoreDB {
		t.Fatalf("expected db selector store, got %q", decoded.PolicyAssets[1].Selector().Store)
	}

	plan := DeletePlan{
		Request:         DeleteRequest{Module: "inventory", DryRun: true, WithPolicy: true, WithRuntime: true, WithFrontend: true, PolicyStore: PolicyStoreCSV},
		Ownership:       ownership,
		Module:          "inventory",
		DryRun:          true,
		Force:           false,
		PolicyStore:     PolicyStoreCSV,
		PolicyStores:    []PolicyStoreKind{PolicyStoreCSV, PolicyStoreDB},
		SourceFiles:     []DeleteItem{{Module: "inventory", Kind: AssetKindSourceFile, Path: "server/modules/inventory/module.go", Origin: AssetOriginGenerated, Managed: true}},
		RuntimeAssets:   []DeleteItem{{Module: "inventory", Kind: AssetKindRuntimeRegistry, Ref: "modules_gen.go", Origin: AssetOriginGenerated, Managed: true}},
		RegistryChanges: []DeleteItem{{Module: "inventory", Kind: AssetKindRuntimeRegistry, Ref: "modules_gen.go", Origin: AssetOriginGenerated, Managed: true}},
		PolicyChanges:   []DeleteItem{{Module: "inventory", Kind: AssetKindPolicyRule, Store: PolicyStoreCSV, Selector: &PolicySelector{Store: PolicyStoreCSV, Module: "inventory", SourceRef: "inventory:view", PType: "p", V0: "inventory", V1: "/api/v1/inventory", V2: "read"}, Origin: AssetOriginGenerated, Managed: true}},
		FrontendChanges: []DeleteItem{{Module: "inventory", Kind: AssetKindFrontendFile, Path: "web/src/views/inventory/index.vue", Origin: AssetOriginGenerated, Managed: true}},
		Warnings:        []string{"legacy module uses compatibility fallback"},
		Conflicts:       []DeleteConflict{{Kind: "shared-path", Severity: "warning", Message: "shared with another module", Path: "web/src/layouts"}},
		Legacy:          ownership.Compatibility,
		Summary:         DeletePlanSummary{SourceFiles: 1, RuntimeAssets: 1, RegistryChanges: 1, PolicyChanges: 1, FrontendChanges: 1, Warnings: 1, Conflicts: 1, Total: 5},
	}

	planPayload, err := json.Marshal(plan)
	if err != nil {
		t.Fatalf("marshal plan: %v", err)
	}
	var decodedPlan DeletePlan
	if err := json.Unmarshal(planPayload, &decodedPlan); err != nil {
		t.Fatalf("unmarshal plan: %v", err)
	}
	if len(decodedPlan.PolicyStores) != 2 {
		t.Fatalf("expected 2 policy stores, got %d", len(decodedPlan.PolicyStores))
	}
	if decodedPlan.Summary.Total != 5 {
		t.Fatalf("expected summary total 5, got %d", decodedPlan.Summary.Total)
	}

	result := DeleteResult{
		Request:    plan.Request,
		Plan:       plan,
		Status:     DeleteStatusSucceeded,
		StartedAt:  time.Unix(1700001000, 0).UTC(),
		FinishedAt: time.Unix(1700001001, 0).UTC(),
		Deleted:    []DeleteItem{{Module: "inventory", Kind: AssetKindSourceFile, Path: "server/modules/inventory/module.go", Origin: AssetOriginGenerated, Managed: true}},
		Skipped:    []DeleteItem{{Module: "inventory", Kind: AssetKindRuntimeMenu, Ref: "system:shared", Origin: AssetOriginShared, Managed: false}},
		Failures:   []DeleteFailure{{Item: DeleteItem{Module: "inventory", Kind: AssetKindPolicyRule, Store: PolicyStoreDB, Selector: &PolicySelector{Store: PolicyStoreDB, Module: "inventory", SourceRef: "inventory:edit", PType: "p", V0: "inventory", V1: "/api/v1/inventory", V2: "update"}}, Reason: "db row locked", Recoverable: true}},
		Warnings:   []string{"refresh enforcer cache manually if required"},
		Summary:    DeleteResultSummary{DeletedSourceFiles: 1, Skipped: 1, Failed: 1, TotalDeleted: 1, ElapsedMillis: 1000},
	}

	resultPayload, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal result: %v", err)
	}
	var decodedResult DeleteResult
	if err := json.Unmarshal(resultPayload, &decodedResult); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if decodedResult.Status != DeleteStatusSucceeded {
		t.Fatalf("expected succeeded status, got %q", decodedResult.Status)
	}
	if decodedResult.Summary.TotalDeleted != 1 {
		t.Fatalf("expected total deleted 1, got %d", decodedResult.Summary.TotalDeleted)
	}
}

func TestLegacyCompatibilityRuleSemantics(t *testing.T) {
	if !(LegacyCompatibilityRule{Mode: LegacyCompatibilityModeConservative}).IsPreviewOnly() {
		t.Fatalf("conservative mode should be preview only")
	}
	if (LegacyCompatibilityRule{Mode: LegacyCompatibilityModeConservative}).AllowsExecution() {
		t.Fatalf("conservative mode should not allow execution")
	}
	if !(LegacyCompatibilityRule{Mode: LegacyCompatibilityModeTemplateInfer}).AllowsExecution() {
		t.Fatalf("template infer mode should allow execution")
	}
}
