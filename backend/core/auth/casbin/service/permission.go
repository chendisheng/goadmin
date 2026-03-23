package service

import (
	"fmt"
	"strings"

	casbinadapter "goadmin/core/auth/casbin/adapter"
	casbinenforcer "goadmin/core/auth/casbin/enforcer"
	coreauthjwt "goadmin/core/auth/jwt"
)

type Authorizer interface {
	EnforceClaims(claims *coreauthjwt.Claims, obj, act string) (bool, error)
}

type Config struct {
	Enabled    bool
	ModelPath  string
	PolicyPath string
}

type PermissionService struct {
	enforcer *casbinenforcer.Enforcer
}

func NewPermissionService(cfg Config) (*PermissionService, error) {
	if !cfg.Enabled {
		return &PermissionService{}, nil
	}
	modelPath := strings.TrimSpace(cfg.ModelPath)
	if modelPath == "" {
		modelPath = "core/auth/casbin/model/rbac.conf"
	}
	policyPath := strings.TrimSpace(cfg.PolicyPath)
	if policyPath == "" {
		policyPath = "core/auth/casbin/adapter/policy.csv"
	}

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

	return &PermissionService{enforcer: enforcer}, nil
}

func (s *PermissionService) String() string {
	if s == nil || s.enforcer == nil {
		return "PermissionService{enabled:false}"
	}
	return fmt.Sprintf("PermissionService{enabled:true, model:%s}", s.enforcer.ModelPath())
}
