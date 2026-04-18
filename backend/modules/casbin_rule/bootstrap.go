// codegen:begin
package casbin_rule

import (
	"fmt"

	corebootstrapcontract "goadmin/core/bootstrap/contract"
	coretransport "goadmin/core/transport"
	casbin_ruleservice "goadmin/modules/casbin_rule/application/service"
	casbin_rulerepo "goadmin/modules/casbin_rule/infrastructure/repo"
	casbin_rulehttp "goadmin/modules/casbin_rule/transport/http"
	"gorm.io/gorm"
)

type Bootstrap struct{}

func NewBootstrap() corebootstrapcontract.Module {
	return Bootstrap{}
}

func (Bootstrap) Name() string {
	return Name
}

func (Bootstrap) ManifestPath() string {
	return ManifestPath
}

func (Bootstrap) Migrate(db *gorm.DB) error {
	return casbin_rulerepo.Migrate(db)
}

func (Bootstrap) Register(group coretransport.RouteRegistrar, deps corebootstrapcontract.Dependencies) error {
	if group == nil {
		return fmt.Errorf("casbin_rule bootstrap requires route registrar")
	}
	if deps.DB == nil {
		return fmt.Errorf("casbin_rule bootstrap requires db")
	}
	repo, err := casbin_rulerepo.NewGormRepository(deps.DB)
	if err != nil {
		return err
	}
	service, err := casbin_ruleservice.New(repo)
	if err != nil {
		return err
	}
	casbin_rulehttp.Register(group, casbin_rulehttp.Dependencies{
		Service: service,
		Logger:  deps.Logger,
	})
	return nil
}

// codegen:end
