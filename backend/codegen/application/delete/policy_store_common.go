package deleteapp

import (
	"context"
	"fmt"
	"strings"
	"time"

	lifecycle "goadmin/codegen/model/lifecycle"
)

func buildPolicyCleanupPreview(req PolicyCleanupRequest, store lifecycle.PolicyStoreKind, rules []lifecycle.PolicyAsset) PolicyCleanupPreview {
	preview := PolicyCleanupPreview{
		Request:  req,
		Store:    store,
		Items:    make([]PolicyCleanupItem, 0, len(req.Selectors)),
		Warnings: make([]string, 0, 2),
	}
	matched := false
	for _, selector := range req.Selectors {
		item := previewPolicySelector(selector, req.Module, rules)
		preview.Items = append(preview.Items, item)
		if item.Decision == "delete" {
			matched = true
		}
		if item.Decision == "conflict" {
			preview.Conflicts = append(preview.Conflicts, lifecycle.DeleteConflict{
				Kind:     "policy-selector-conflict",
				Severity: conflictSeverityHigh,
				Message:  item.Reason,
				Ref:      item.Rule.SourceRef,
				Metadata: map[string]any{"selector": selectorToSummary(selector)},
			})
		}
	}
	if len(preview.Items) == 0 {
		preview.Warnings = append(preview.Warnings, "no policy candidates matched the request")
	}
	if !matched {
		preview.Warnings = append(preview.Warnings, "no policy rules were selected for deletion")
	}
	preview.Summary = decisionSummary(preview.Items)
	preview.Summary.Backend = string(store)
	preview.Summary.Validated = true
	return preview
}

func newPolicyCleanupResult(preview PolicyCleanupPreview, started time.Time) PolicyCleanupResult {
	return PolicyCleanupResult{
		Preview: preview,
		Audit: PolicyCleanupAudit{
			StartedAt:  started,
			Backend:    string(preview.Store),
			Operation:  "delete",
			Validation: "pending",
		},
	}
}

func finalizePolicyCleanupResult(result *PolicyCleanupResult, preview PolicyCleanupPreview, deleted, skipped []PolicyCleanupItem, verified bool, validation string, cacheRefreshed bool) {
	if result == nil {
		return
	}
	result.Deleted = deleted
	result.Skipped = skipped
	result.Verified = verified
	result.Summary = decisionSummary(preview.Items)
	result.Summary.Backend = string(preview.Store)
	result.Summary.Validated = true
	result.Audit.FinishedAt = nowUTC()
	result.Audit.Validation = validation
	result.Audit.CacheRefreshed = cacheRefreshed
}

func validatePolicyStoreRules(ctx context.Context, load func(context.Context) ([]lifecycle.PolicyAsset, error)) error {
	if load == nil {
		return fmt.Errorf("policy store loader is not configured")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	_, err := load(ctx)
	return err
}

func policyAssetsByModule(assets []lifecycle.PolicyAsset, module string) []lifecycle.PolicyAsset {
	module = strings.TrimSpace(module)
	if module == "" || len(assets) == 0 {
		return assets
	}
	filtered := make([]lifecycle.PolicyAsset, 0, len(assets))
	for _, asset := range assets {
		assetModule := policyAssetModule(asset)
		if assetModule == "" || strings.EqualFold(assetModule, module) {
			filtered = append(filtered, asset)
		}
	}
	if len(filtered) == 0 {
		return assets
	}
	return filtered
}

func policyAssetModule(asset lifecycle.PolicyAsset) string {
	if module := strings.TrimSpace(asset.Module); module != "" {
		return module
	}
	if asset.Metadata != nil {
		if value, ok := asset.Metadata["module"].(string); ok {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func policyAssetSourceRef(asset lifecycle.PolicyAsset) string {
	if ref := strings.TrimSpace(asset.SourceRef); ref != "" {
		return ref
	}
	if asset.Metadata != nil {
		if value, ok := asset.Metadata["source_ref"].(string); ok {
			return strings.TrimSpace(value)
		}
	}
	return strings.TrimSpace(selectorToSummary(asset.Selector()))
}

func policyAssetManaged(asset lifecycle.PolicyAsset) bool {
	if asset.Managed {
		return true
	}
	if asset.Metadata == nil {
		return false
	}
	value, ok := asset.Metadata["managed"]
	if !ok {
		return false
	}
	switch typed := value.(type) {
	case bool:
		return typed
	case string:
		return strings.EqualFold(strings.TrimSpace(typed), "true")
	default:
		return false
	}
}

func previewPolicySelector(selector lifecycle.PolicySelector, module string, rules []lifecycle.PolicyAsset) PolicyCleanupItem {
	item := PolicyCleanupItem{Selector: selector, Rule: selectorToPolicyAsset(selector, selector.SourceRef), Reason: "selector did not match any policy", Decision: "skip"}
	if !selectorManaged(selector) {
		item.Reason = "selector is not marked as managed"
		return item
	}
	if !selectorMatchesModule(selector, module) {
		item.Reason = "selector module does not match request module"
		return item
	}
	matches := make([]lifecycle.PolicyAsset, 0)
	for _, rule := range rules {
		if ruleMatchesSelector(rule, selector) {
			matches = append(matches, rule)
		}
	}
	if len(matches) == 0 {
		item.Reason = "no matching policy rule found"
		return item
	}
	item.Decision = "delete"
	item.MatchCount = len(matches)
	item.Rule = matches[0]
	item.Reason = "matched by structured selector"
	return item
}

func findMatchingSelector(req PolicyCleanupRequest, rule lifecycle.PolicyAsset) *lifecycle.PolicySelector {
	for i := range req.Selectors {
		selector := req.Selectors[i]
		if !selectorManaged(selector) {
			continue
		}
		if !selectorMatchesModule(selector, req.Module) {
			continue
		}
		if ruleMatchesSelector(rule, selector) {
			copySelector := selector
			return &copySelector
		}
	}
	return nil
}

func verifyPolicySelections(rules []lifecycle.PolicyAsset, selectors []lifecycle.PolicySelector) bool {
	for _, selector := range selectors {
		if !selectorManaged(selector) {
			continue
		}
		for _, rule := range rules {
			if ruleMatchesSelector(rule, selector) {
				return false
			}
		}
	}
	return true
}

func collectSkippedItems(items []PolicyCleanupItem, deleted []PolicyCleanupItem) []PolicyCleanupItem {
	if len(items) == 0 {
		return nil
	}
	deletedKeys := make(map[string]struct{}, len(deleted))
	for _, item := range deleted {
		deletedKeys[policySelectorKey(item.Selector)] = struct{}{}
	}
	skipped := make([]PolicyCleanupItem, 0)
	for _, item := range items {
		if item.Decision != "delete" {
			skipped = append(skipped, item)
			continue
		}
		if _, ok := deletedKeys[policySelectorKey(item.Selector)]; ok {
			continue
		}
		skipped = append(skipped, item)
	}
	return skipped
}

func policySelectorKey(selector lifecycle.PolicySelector) string {
	parts := []string{
		string(choosePolicyStoreKind(selector.Store, lifecycle.PolicyStoreUnknown)),
		strings.TrimSpace(selector.Module),
		strings.TrimSpace(selector.SourceRef),
		strings.TrimSpace(selector.PType),
		strings.TrimSpace(selector.V0),
		strings.TrimSpace(selector.V1),
		strings.TrimSpace(selector.V2),
		strings.TrimSpace(selector.V3),
		strings.TrimSpace(selector.V4),
		strings.TrimSpace(selector.V5),
	}
	return strings.Join(parts, "\x1f")
}

func selectorToSummary(selector lifecycle.PolicySelector) string {
	parts := []string{
		strings.TrimSpace(selector.PType),
		strings.TrimSpace(selector.V0),
		strings.TrimSpace(selector.V1),
		strings.TrimSpace(selector.V2),
		strings.TrimSpace(selector.V3),
		strings.TrimSpace(selector.V4),
		strings.TrimSpace(selector.V5),
	}
	filtered := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			filtered = append(filtered, part)
		}
	}
	if len(filtered) == 0 {
		return ""
	}
	return strings.Join(filtered, " ")
}

func selectorManaged(selector lifecycle.PolicySelector) bool {
	if value, ok := selector.Metadata["managed"]; ok {
		switch typed := value.(type) {
		case bool:
			return typed
		case string:
			return strings.EqualFold(strings.TrimSpace(typed), "true")
		}
	}
	return false
}

func selectorModule(selector lifecycle.PolicySelector) string {
	if module := strings.TrimSpace(selector.Module); module != "" {
		return module
	}
	if value, ok := selector.Metadata["module"].(string); ok {
		return strings.TrimSpace(value)
	}
	return ""
}

func selectorMatchesModule(selector lifecycle.PolicySelector, module string) bool {
	module = strings.TrimSpace(module)
	if module == "" {
		return true
	}
	selected := selectorModule(selector)
	return selected == "" || strings.EqualFold(selected, module)
}

func selectorToPolicyAsset(selector lifecycle.PolicySelector, sourceRef string) lifecycle.PolicyAsset {
	return lifecycle.PolicyAsset{
		Store:     selector.Store,
		Module:    strings.TrimSpace(selector.Module),
		SourceRef: strings.TrimSpace(sourceRef),
		PType:     strings.TrimSpace(selector.PType),
		V0:        strings.TrimSpace(selector.V0),
		V1:        strings.TrimSpace(selector.V1),
		V2:        strings.TrimSpace(selector.V2),
		V3:        strings.TrimSpace(selector.V3),
		V4:        strings.TrimSpace(selector.V4),
		V5:        strings.TrimSpace(selector.V5),
		Metadata:  cloneAnyMap(selector.Metadata),
	}
}

func ruleMatchesSelector(rule lifecycle.PolicyAsset, selector lifecycle.PolicySelector) bool {
	if selector.Store.IsKnown() && rule.Store.IsKnown() && selector.Store != rule.Store {
		return false
	}
	if strings.TrimSpace(selector.PType) != "" && !strings.EqualFold(strings.TrimSpace(selector.PType), strings.TrimSpace(rule.PType)) {
		return false
	}
	selectorValues := selector.Values()
	ruleValues := rule.Selector().Values()
	for i := range selectorValues {
		if strings.TrimSpace(selectorValues[i]) == "" {
			continue
		}
		if strings.TrimSpace(selectorValues[i]) != strings.TrimSpace(ruleValues[i]) {
			return false
		}
	}
	return true
}

func decisionSummary(items []PolicyCleanupItem) PolicyCleanupSummary {
	summary := PolicyCleanupSummary{Total: len(items)}
	for _, item := range items {
		switch item.Decision {
		case "delete":
			summary.Selected++
			summary.Deleted++
		case "skip":
			summary.Skipped++
		case "conflict":
			summary.Conflicts++
		case "fail":
			summary.Failures++
		}
		if item.MatchCount > 0 {
			summary.Validated = true
		}
	}
	return summary
}

func nowUTC() time.Time {
	return time.Now().UTC()
}
