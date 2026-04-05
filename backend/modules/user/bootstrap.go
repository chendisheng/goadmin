package user

import (
	"fmt"

	corebootstrapcontract "goadmin/core/bootstrap/contract"
	coreevent "goadmin/core/event"
	coretransport "goadmin/core/transport"
	userservice "goadmin/modules/user/application/service"
	userrepo "goadmin/modules/user/infrastructure/repo"
	userhttp "goadmin/modules/user/transport/http"

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
	return userrepo.Migrate(db)
}

func (Bootstrap) Register(group coretransport.RouteRegistrar, deps corebootstrapcontract.Dependencies) error {
	if group == nil {
		return fmt.Errorf("user bootstrap requires route registrar")
	}
	if deps.DB == nil {
		return fmt.Errorf("user bootstrap requires db")
	}
	repo, err := userrepo.NewGormRepository(deps.DB)
	if err != nil {
		return err
	}
	var bus coreevent.Bus = deps.EventBus
	service, err := userservice.New(repo, bus)
	if err != nil {
		return err
	}
	userhttp.Register(group, userhttp.Dependencies{Service: service, Logger: deps.Logger})
	return nil
}
