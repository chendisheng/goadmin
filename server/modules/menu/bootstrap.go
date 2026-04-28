package menu

import (
	"fmt"

	corebootstrapcontract "goadmin/core/bootstrap/contract"
	coretransport "goadmin/core/transport"
	menurepo "goadmin/modules/menu/infrastructure/repo"

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
	return menurepo.Migrate(db)
}

func (Bootstrap) Register(group coretransport.RouteRegistrar, deps corebootstrapcontract.Dependencies) error {
	if group == nil {
		return fmt.Errorf("menu bootstrap requires route registrar")
	}
	if deps.DB == nil {
		return fmt.Errorf("menu bootstrap requires db")
	}
	return nil
}
