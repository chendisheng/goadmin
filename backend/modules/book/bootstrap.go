// codegen:begin
package book

import (
	"fmt"

	corebootstrapcontract "goadmin/core/bootstrap/contract"
	coretransport "goadmin/core/transport"
	bookservice "goadmin/modules/book/application/service"
	bookrepo "goadmin/modules/book/infrastructure/repo"
	bookhttp "goadmin/modules/book/transport/http"

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
	return bookrepo.Migrate(db)
}

func (Bootstrap) Register(group coretransport.RouteRegistrar, deps corebootstrapcontract.Dependencies) error {
	if group == nil {
		return fmt.Errorf("book bootstrap requires route registrar")
	}
	if deps.DB == nil {
		return fmt.Errorf("book bootstrap requires db")
	}
	repo, err := bookrepo.NewGormRepository(deps.DB)
	if err != nil {
		return err
	}
	service, err := bookservice.New(repo)
	if err != nil {
		return err
	}
	bookhttp.Register(group, bookhttp.Dependencies{Service: service, Logger: deps.Logger})
	return nil
}

// codegen:end
