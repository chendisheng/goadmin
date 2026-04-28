package deleteapp

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	lifecycle "goadmin/codegen/model/lifecycle"
	"os"
	"path/filepath"
	"strings"
)

type CSVPolicyStore struct {
	path string
}

func NewCSVPolicyStore(path string) (*CSVPolicyStore, error) {
	clean := strings.TrimSpace(path)
	if clean == "" {
		return nil, fmt.Errorf("csv policy path is required")
	}
	return &CSVPolicyStore{path: clean}, nil
}

func (s *CSVPolicyStore) Kind() lifecycle.PolicyStoreKind {
	return lifecycle.PolicyStoreCSV
}

func (s *CSVPolicyStore) ListByModule(ctx context.Context, module string) ([]lifecycle.PolicyAsset, error) {
	if s == nil {
		return nil, fmt.Errorf("csv policy store is not configured")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	rules, err := loadCSVRules(s.path)
	if err != nil {
		return nil, err
	}
	return policyAssetsByModule(rules, module), nil
}

func (s *CSVPolicyStore) DeleteBySelector(ctx context.Context, selector lifecycle.PolicySelector) (int, error) {
	if s == nil {
		return 0, fmt.Errorf("csv policy store is not configured")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	rules, err := loadCSVRules(s.path)
	if err != nil {
		return 0, err
	}
	filtered := make([]lifecycle.PolicyAsset, 0, len(rules))
	deleted := 0
	for _, rule := range rules {
		if ruleMatchesSelector(rule, selector) {
			deleted++
			continue
		}
		filtered = append(filtered, rule)
	}
	if deleted == 0 {
		return 0, nil
	}
	if err := writeCSVRulesAtomic(s.path, filtered); err != nil {
		return 0, err
	}
	return deleted, nil
}

func (s *CSVPolicyStore) Validate(ctx context.Context) error {
	if s == nil {
		return fmt.Errorf("csv policy store is not configured")
	}
	return validatePolicyStoreRules(ctx, func(ctx context.Context) ([]lifecycle.PolicyAsset, error) {
		_ = ctx
		return loadCSVRules(s.path)
	})
}

func (s *CSVPolicyStore) Preview(ctx context.Context, req PolicyCleanupRequest) (PolicyCleanupPreview, error) {
	if s == nil {
		return PolicyCleanupPreview{}, fmt.Errorf("csv policy store is not configured")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	req = req.Normalize(s.Kind())
	if err := req.ValidateRequest(); err != nil {
		return PolicyCleanupPreview{}, err
	}
	rules, err := s.ListByModule(ctx, req.Module)
	if err != nil {
		return PolicyCleanupPreview{}, err
	}
	return buildPolicyCleanupPreview(req, s.Kind(), rules), nil
}

func (s *CSVPolicyStore) Delete(ctx context.Context, req PolicyCleanupRequest) (PolicyCleanupResult, error) {
	if s == nil {
		return PolicyCleanupResult{}, fmt.Errorf("csv policy store is not configured")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	req = req.Normalize(s.Kind())
	if err := req.ValidateRequest(); err != nil {
		return PolicyCleanupResult{}, err
	}
	started := nowUTC()
	preview, err := s.Preview(ctx, req)
	if err != nil {
		return PolicyCleanupResult{}, err
	}
	result := newPolicyCleanupResult(preview, started)
	if len(preview.Conflicts) > 0 {
		result.Warnings = append(result.Warnings, "policy cleanup aborted because preview contained conflicts")
		result.Audit.FinishedAt = nowUTC()
		result.Audit.Validation = "blocked"
		return result, fmt.Errorf("policy cleanup preview has conflicts")
	}
	if len(preview.Items) == 0 {
		finalizePolicyCleanupResult(&result, preview, nil, nil, true, "no-op", false)
		return result, nil
	}
	rules, err := loadCSVRules(s.path)
	if err != nil {
		return result, err
	}
	filtered := make([]lifecycle.PolicyAsset, 0, len(rules))
	deleted := make([]PolicyCleanupItem, 0, len(preview.Items))
	for _, rule := range rules {
		matchedSelector := findMatchingSelector(req, rule)
		if matchedSelector == nil {
			filtered = append(filtered, rule)
			continue
		}
		deleted = append(deleted, PolicyCleanupItem{
			Selector:   *matchedSelector,
			Rule:       rule,
			Decision:   "delete",
			Reason:     "matched by selector",
			MatchCount: 1,
		})
	}
	if err := writeCSVRulesAtomic(s.path, filtered); err != nil {
		return result, err
	}
	verifiedRules, err := loadCSVRules(s.path)
	if err != nil {
		return result, err
	}
	verified := verifyPolicySelections(verifiedRules, req.Selectors)
	if verified {
		result.Verified = true
	}
	finalizePolicyCleanupResult(&result, preview, deleted, collectSkippedItems(preview.Items, deleted), verified, "ok", false)
	return result, nil
}

func loadCSVRules(path string) ([]lifecycle.PolicyAsset, error) {
	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("open casbin policy file: %w", err)
	}
	defer file.Close()
	var rules []lifecycle.PolicyAsset
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}
		parts := splitCSVLine(line)
		rule, err := normalizeCSVRule(parts)
		if err != nil {
			return nil, fmt.Errorf("invalid casbin policy line %q: %w", line, err)
		}
		rule.SourceRef = line
		rules = append(rules, rule)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan casbin policy file: %w", err)
	}
	return rules, nil
}

func writeCSVRulesAtomic(path string, rules []lifecycle.PolicyAsset) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create casbin policy directory: %w", err)
	}
	tmp := path + ".tmp"
	file, err := os.Create(tmp)
	if err != nil {
		return fmt.Errorf("create casbin policy temp file: %w", err)
	}
	writer := bufio.NewWriter(file)
	for _, rule := range rules {
		if _, err := writer.WriteString(formatCSVRuleLine(rule) + "\n"); err != nil {
			_ = file.Close()
			_ = os.Remove(tmp)
			return fmt.Errorf("write casbin policy file: %w", err)
		}
	}
	if err := writer.Flush(); err != nil {
		_ = file.Close()
		_ = os.Remove(tmp)
		return fmt.Errorf("flush casbin policy file: %w", err)
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("close casbin policy temp file: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("replace casbin policy file: %w", err)
	}
	return nil
}

func splitCSVLine(line string) []string {
	parts := strings.Split(line, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

func normalizeCSVRule(parts []string) (lifecycle.PolicyAsset, error) {
	trimmed := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed = append(trimmed, strings.TrimSpace(part))
	}
	if len(trimmed) >= 4 && strings.EqualFold(trimmed[0], "p") {
		trimmed = trimmed[1:]
	}
	if len(trimmed) < 3 || len(trimmed) > 6 {
		return lifecycle.PolicyAsset{}, fmt.Errorf("invalid policy row")
	}
	asset := lifecycle.PolicyAsset{Store: lifecycle.PolicyStoreCSV, PType: "p", V0: trimmed[0], V1: trimmed[1], V2: trimmed[2]}
	if len(trimmed) > 3 {
		asset.V3 = trimmed[3]
	}
	if len(trimmed) > 4 {
		asset.V4 = trimmed[4]
	}
	if len(trimmed) > 5 {
		asset.V5 = trimmed[5]
	}
	return asset, nil
}

func formatCSVRuleLine(rule lifecycle.PolicyAsset) string {
	values := []string{
		strings.TrimSpace(rule.V0),
		strings.TrimSpace(rule.V1),
		strings.TrimSpace(rule.V2),
		strings.TrimSpace(rule.V3),
		strings.TrimSpace(rule.V4),
		strings.TrimSpace(rule.V5),
	}
	last := len(values) - 1
	for last >= 2 && values[last] == "" {
		last--
	}
	parts := append([]string{"p"}, values[:last+1]...)
	return strings.Join(parts, ", ")
}
