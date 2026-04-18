// codegen:begin
package casbin_model

import (
	"fmt"

	corebootstrapcontract "goadmin/core/bootstrap/contract"
	coretransport "goadmin/core/transport"
	casbin_modelservice "goadmin/modules/casbin_model/application/service"
	casbin_modelrepo "goadmin/modules/casbin_model/infrastructure/repo"
	casbin_modelhttp "goadmin/modules/casbin_model/transport/http"
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
	return casbin_modelrepo.Migrate(db)
}

func (Bootstrap) Register(group coretransport.RouteRegistrar, deps corebootstrapcontract.Dependencies) error {
	if group == nil {
		return fmt.Errorf("casbin_model bootstrap requires route registrar")
	}
	if deps.DB == nil {
		return fmt.Errorf("casbin_model bootstrap requires db")
	}
	repo, err := casbin_modelrepo.NewGormRepository(deps.DB)
	if err != nil {
		return err
	}
	service, err := casbin_modelservice.New(repo)
	if err != nil {
		return err
	}
	casbin_modelhttp.Register(group, casbin_modelhttp.Dependencies{
		Service: service,
		Logger:  deps.Logger,
	})
	return nil
}

// codegen:end
