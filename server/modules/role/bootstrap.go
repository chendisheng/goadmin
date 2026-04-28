package role

import (
	"fmt"

	corebootstrapcontract "goadmin/core/bootstrap/contract"
	coretransport "goadmin/core/transport"
	roleservice "goadmin/modules/role/application/service"
	rolerepo "goadmin/modules/role/infrastructure/repo"
	rolehttp "goadmin/modules/role/transport/http"

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
	return rolerepo.Migrate(db)
}

func (Bootstrap) Register(group coretransport.RouteRegistrar, deps corebootstrapcontract.Dependencies) error {
	if group == nil {
		return fmt.Errorf("role bootstrap requires route registrar")
	}
	if deps.DB == nil {
		return fmt.Errorf("role bootstrap requires db")
	}
	repo, err := rolerepo.NewGormRepository(deps.DB)
	if err != nil {
		return err
	}
	service, err := roleservice.New(repo)
	if err != nil {
		return err
	}
	rolehttp.Register(group, rolehttp.Dependencies{Service: service, Logger: deps.Logger})
	return nil
}
