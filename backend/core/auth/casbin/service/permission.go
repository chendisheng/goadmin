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
		if err := seedDBPolicies(store, policyPath); err != nil {
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

func seedDBPolicies(store casbinadapter.GormPolicyStore, policyPath string) error {
	if store == nil {
		return fmt.Errorf("casbin policy store is required")
	}
	rules, err := store.LoadPolicies(context.Background())
	if err != nil {
		return err
	}
	if len(rules) > 0 {
		return nil
	}
	fileAdapter, err := casbinadapter.NewFileAdapter(policyPath)
	if err != nil {
		return err
	}
	rules, err = fileAdapter.LoadRules()
	if err != nil {
		return err
	}
	if len(rules) == 0 {
		return nil
	}
	return store.SavePolicies(context.Background(), rules)
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
