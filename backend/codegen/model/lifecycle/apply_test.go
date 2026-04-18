package lifecycle

import (
	"encoding/json"
	"testing"
	"time"
)

func TestApplyRequestNormalizeAndValidate(t *testing.T) {
	req := ApplyRequest{
		Module:      "  inventory  ",
		Kind:        "  crud  ",
		PolicyStore: NormalizePolicyStoreKind(" db "),
		Execution:   ApplyExecutionRule{Notes: []string{"  keep generated only  ", "", "keep generated only"}},
	}

	normalized := req.Normalize()
	if normalized.Module != "inventory" {
		t.Fatalf("expected normalized module inventory, got %q", normalized.Module)
	}
	if normalized.Kind != "crud" {
		t.Fatalf("expected normalized kind crud, got %q", normalized.Kind)
	}
	if normalized.PolicyStore != PolicyStoreDB {
		t.Fatalf("expected db policy store, got %q", normalized.PolicyStore)
	}
	if len(normalized.Execution.Notes) != 1 || normalized.Execution.Notes[0] != "keep generated only" {
		t.Fatalf("expected deduplicated normalized execution notes, got %#v", normalized.Execution.Notes)
	}
	if err := normalized.Validate(); err != nil {
		t.Fatalf("expected request to validate, got error: %v", err)
	}
}

func TestApplyPlanAndResultRoundTrip(t *testing.T) {
	pairing := DefaultApplyDeletePairing()
	plan := ApplyPlan{
		Request: ApplyRequest{
			Module:      "inventory",
			DryRun:      true,
			Install:     true,
			Refresh:     true,
			Verify:      true,
			PolicyStore: PolicyStoreCSV,
		},
		Module:       "inventory",
		Kind:         "crud",
		DryRun:       true,
		Force:        false,
		PolicyStore:  PolicyStoreCSV,
		Generation:   []ApplyItem{{Module: "inventory", Kind: "source-file", Path: "backend/modules/inventory/module.go", Stage: ApplyStageGenerate, Action: "created", Managed: true}},
		Installation: []ApplyItem{{Module: "inventory", Kind: "runtime-menu", Path: "/system/inventory", Stage: ApplyStageInstall, Action: "installed", Managed: true}},
		PolicySync:   []ApplyItem{{Module: "inventory", Kind: "policy-rule", Ref: "inventory:view", Store: PolicyStoreCSV, Stage: ApplyStagePolicySync, Action: "synced", Managed: true}},
		Validation:   []ApplyValidationIssue{{Item: ApplyItem{Module: "inventory", Kind: "runtime-menu", Path: "/system/inventory", Stage: ApplyStageValidate}, Stage: ApplyStageValidate, Message: "menu exists"}},
		Warnings:     []string{"shared resources are skipped"},
		Conflicts:    []ApplyConflict{{Kind: "shared-resource", Severity: "warning", Message: "shared with another module", Path: "/system/shared"}},
		Pairing:      pairing,
		Summary:      ApplyPlanSummary{Generation: 1, Installation: 1, PolicySync: 1, Validation: 1, Warnings: 1, Conflicts: 1, Total: 3},
	}

	planPayload, err := json.Marshal(plan)
	if err != nil {
		t.Fatalf("marshal plan: %v", err)
	}
	var decodedPlan ApplyPlan
	if err := json.Unmarshal(planPayload, &decodedPlan); err != nil {
		t.Fatalf("unmarshal plan: %v", err)
	}
	if decodedPlan.Pairing.Operation != "apply" || decodedPlan.Pairing.Counterpart != "delete" {
		t.Fatalf("expected default pairing to survive json round trip, got %#v", decodedPlan.Pairing)
	}
	if decodedPlan.Summary.Total != 3 {
		t.Fatalf("expected plan summary total 3, got %d", decodedPlan.Summary.Total)
	}

	result := ApplyResult{
		Request:    plan.Request,
		Plan:       plan,
		Status:     ApplyStatusSucceeded,
		StartedAt:  time.Unix(1700001000, 0).UTC(),
		FinishedAt: time.Unix(1700001001, 0).UTC(),
		Generation: ApplyGenerationReport{
			StartedAt:  time.Unix(1700001000, 0).UTC(),
			FinishedAt: time.Unix(1700001000, 0).UTC(),
			Status:     "succeeded",
			Created:    []ApplyItem{{Module: "inventory", Kind: "source-file", Path: "backend/modules/inventory/module.go", Stage: ApplyStageGenerate, Action: "created", Managed: true}},
		},
		Installation: ApplyInstallationReport{
			StartedAt:  time.Unix(1700001000, 0).UTC(),
			FinishedAt: time.Unix(1700001000, 0).UTC(),
			Status:     "succeeded",
			Created:    []ApplyItem{{Module: "inventory", Kind: "runtime-menu", Path: "/system/inventory", Stage: ApplyStageInstall, Action: "installed", Managed: true}},
		},
		PolicySync: ApplyPolicySyncReport{
			StartedAt:  time.Unix(1700001000, 0).UTC(),
			FinishedAt: time.Unix(1700001000, 0).UTC(),
			Status:     "succeeded",
			Store:      PolicyStoreCSV,
			Synced:     []ApplyItem{{Module: "inventory", Kind: "policy-rule", Ref: "inventory:view", Store: PolicyStoreCSV, Stage: ApplyStagePolicySync, Action: "synced", Managed: true}},
		},
		Validation: ApplyValidationReport{
			StartedAt:  time.Unix(1700001000, 0).UTC(),
			FinishedAt: time.Unix(1700001001, 0).UTC(),
			Status:     "passed",
			Verified:   true,
			Checked:    1,
			Issues:     []ApplyValidationIssue{{Stage: ApplyStageValidate, Message: "all good"}},
		},
		Warnings: []string{"refresh recommended"},
		Pairing:  pairing,
		Summary:  ApplyResultSummary{Generated: 1, Installed: 1, PolicySynced: 1, ValidationChecks: 1, ValidationIssues: 0, TotalApplied: 3, ElapsedMillis: 1000},
	}

	resultPayload, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal result: %v", err)
	}
	var decodedResult ApplyResult
	if err := json.Unmarshal(resultPayload, &decodedResult); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}
	if decodedResult.Status != ApplyStatusSucceeded {
		t.Fatalf("expected succeeded status, got %q", decodedResult.Status)
	}
	if decodedResult.Summary.TotalApplied != 3 {
		t.Fatalf("expected total applied 3, got %d", decodedResult.Summary.TotalApplied)
	}
	if !decodedResult.Validation.Verified {
		t.Fatalf("expected validation to be verified")
	}
}

func TestApplyDeletePairingDefaults(t *testing.T) {
	pairing := DefaultApplyDeletePairing()
	if pairing.Operation != "apply" {
		t.Fatalf("expected apply operation, got %q", pairing.Operation)
	}
	if pairing.Counterpart != "delete" {
		t.Fatalf("expected delete counterpart, got %q", pairing.Counterpart)
	}
	if len(pairing.NonSymmetricNotes) == 0 {
		t.Fatalf("expected non symmetric notes to be populated")
	}
}
