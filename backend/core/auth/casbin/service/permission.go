package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	casbinadapter "goadmin/core/auth/casbin/adapter"
	casbinenforcer "goadmin/core/auth/casbin/enforcer"
	coreauthjwt "goadmin/core/auth/jwt"

	"gorm.io/gorm"
)

type Authorizer interface {
	EnforceClaims(claims *coreauthjwt.Claims, obj, act string) (bool, error)
}

type Config struct {
	Enabled    bool
	Source     string
	DB         *gorm.DB
	ModelPath  string
	PolicyPath string
}

type PermissionService struct {
	enforcer *casbinenforcer.Enforcer
	source   string
	model    string
	policy   string
}

func NewPermissionService(cfg Config) (*PermissionService, error) {
	if !cfg.Enabled {
		return &PermissionService{}, nil
	}
	source := strings.ToLower(strings.TrimSpace(cfg.Source))
	if source == "" {
		source = "file"
	}
	modelPath := strings.TrimSpace(cfg.ModelPath)
	if modelPath == "" {
		modelPath = "core/auth/casbin/model/rbac.conf"
	}
	policyPath := strings.TrimSpace(cfg.PolicyPath)
	if policyPath == "" {
		policyPath = "core/auth/casbin/adapter/policy.csv"
	}

	switch source {
	case "file":
		fileAdapter, err := casbinadapter.NewFileAdapter(policyPath)
		if err != nil {
			return nil, err
		}
		enforcer, err := casbinenforcer.New(casbinenforcer.Config{
			ModelPath: modelPath,
			Adapter:   fileAdapter,
		})
		if err != nil {
			return nil, err
		}
		return &PermissionService{enforcer: enforcer, source: source, model: modelPath, policy: policyPath}, nil
	case "db":
		if cfg.DB == nil {
			return nil, fmt.Errorf("casbin source db requires db")
		}
		if err := casbinadapter.Migrate(cfg.DB); err != nil {
			return nil, err
		}
		store, err := casbinadapter.NewGormStore(cfg.DB)
		if err != nil {
			return nil, err
		}
		if err := seedDBModel(store, modelPath); err != nil {
			return nil, err
		}
		if err := syncDBPolicies(store, policyPath); err != nil {
			return nil, err
		}
		modelContent, err := store.LoadModel(context.Background(), filepath.Base(modelPath))
		if err != nil {
			return nil, err
		}
		tempModelPath, err := writeTempModel(modelContent)
		if err != nil {
			return nil, err
		}
		defer func() {
			_ = os.Remove(tempModelPath)
		}()
		policyAdapter, err := casbinadapter.NewGormAdapter(store)
		if err != nil {
			return nil, err
		}
		enforcer, err := casbinenforcer.New(casbinenforcer.Config{
			ModelPath: tempModelPath,
			Adapter:   policyAdapter,
		})
		if err != nil {
			return nil, err
		}
		return &PermissionService{enforcer: enforcer, source: source, model: tempModelPath, policy: policyPath}, nil
	default:
		return nil, fmt.Errorf("auth.casbin.source must be file or db")
	}
}

func (s *PermissionService) String() string {
	if s == nil || s.enforcer == nil {
		return "PermissionService{enabled:false}"
	}
	return fmt.Sprintf("PermissionService{enabled:true, source:%s, model:%s}", s.source, s.enforcer.ModelPath())
}

func seedDBModel(store *casbinadapter.GormCasbinStore, modelPath string) error {
	if store == nil {
		return fmt.Errorf("casbin model store is required")
	}
	name := filepath.Base(strings.TrimSpace(modelPath))
	content, err := store.LoadModel(context.Background(), name)
	if err != nil {
		return err
	}
	if strings.TrimSpace(content) != "" {
		return nil
	}
	data, err := os.ReadFile(modelPath)
	if err != nil {
		return fmt.Errorf("read casbin model file: %w", err)
	}
	if err := store.SaveModel(context.Background(), name, string(data)); err != nil {
		return err
	}
	return nil
}

func syncDBPolicies(store casbinadapter.GormPolicyStore, policyPath string) error {
	if store == nil {
		return fmt.Errorf("casbin policy store is required")
	}
	fileAdapter, err := casbinadapter.NewFileAdapter(policyPath)
	if err != nil {
		return err
	}
	fileRules, err := fileAdapter.LoadRules()
	if err != nil {
		return err
	}
	existingRules, err := store.LoadPolicies(context.Background())
	if err != nil {
		return err
	}
	mergedRules := mergePolicyRules(defaultRules(), fileRules)
	mergedRules = mergePolicyRules(mergedRules, existingRules)
	if samePolicyRules(existingRules, mergedRules) {
		return nil
	}
	return store.SavePolicies(context.Background(), mergedRules)
}

func mergePolicyRules(primary, fallback []casbinadapter.Rule) []casbinadapter.Rule {
	seen := make(map[string]struct{}, len(primary)+len(fallback))
	merged := make([]casbinadapter.Rule, 0, len(primary)+len(fallback))
	appendRule := func(rule casbinadapter.Rule) {
		key := policyRuleKey(rule)
		if key == "" {
			return
		}
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		merged = append(merged, rule)
	}
	for _, rule := range primary {
		appendRule(rule)
	}
	for _, rule := range fallback {
		appendRule(rule)
	}
	return merged
}

func samePolicyRules(left, right []casbinadapter.Rule) bool {
	if len(left) != len(right) {
		return false
	}
	counts := make(map[string]int, len(left))
	for _, rule := range left {
		key := policyRuleKey(rule)
		if key == "" {
			return false
		}
		counts[key]++
	}
	for _, rule := range right {
		key := policyRuleKey(rule)
		if key == "" {
			return false
		}
		if counts[key] == 0 {
			return false
		}
		counts[key]--
	}
	for _, remaining := range counts {
		if remaining != 0 {
			return false
		}
	}
	return true
}

func policyRuleKey(rule casbinadapter.Rule) string {
	subject := strings.TrimSpace(rule.Subject)
	object := strings.TrimSpace(rule.Object)
	action := strings.TrimSpace(rule.Action)
	if subject == "" || object == "" || action == "" {
		return ""
	}
	return subject + "\x1f" + object + "\x1f" + action
}

func writeTempModel(content string) (string, error) {
	tmp, err := os.CreateTemp("", "goadmin-casbin-model-*.conf")
	if err != nil {
		return "", fmt.Errorf("create casbin model temp file: %w", err)
	}
	defer func() {
		_ = tmp.Close()
	}()
	if _, err := tmp.WriteString(content); err != nil {
		_ = os.Remove(tmp.Name())
		return "", fmt.Errorf("write casbin model temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmp.Name())
		return "", fmt.Errorf("close casbin model temp file: %w", err)
	}
	return tmp.Name(), nil
}
