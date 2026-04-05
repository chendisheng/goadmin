package order

import (
	"fmt"

	corebootstrapcontract "goadmin/core/bootstrap/contract"
	coretransport "goadmin/core/transport"
	orderservice "goadmin/modules/order/application/service"
	orderrepo "goadmin/modules/order/infrastructure/repo"
	orderhttp "goadmin/modules/order/transport/http"

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
	return orderrepo.Migrate(db)
}

func (Bootstrap) Register(group coretransport.RouteRegistrar, deps corebootstrapcontract.Dependencies) error {
	if group == nil {
		return fmt.Errorf("order bootstrap requires route registrar")
	}
	if deps.DB == nil {
		return fmt.Errorf("order bootstrap requires db")
	}
	repo, err := orderrepo.NewGormRepository(deps.DB)
	if err != nil {
		return err
	}
	service, err := orderservice.New(repo)
	if err != nil {
		return err
	}
	orderhttp.Register(group, orderhttp.Dependencies{Service: service, Logger: deps.Logger})
	return nil
}
