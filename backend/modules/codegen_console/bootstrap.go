package codegen_console

import (
	"fmt"

	corebootstrapcontract "goadmin/core/bootstrap/contract"
	coretransport "goadmin/core/transport"
	codegen_consoleservice "goadmin/modules/codegen_console/application/service"
	codegen_consolerepo "goadmin/modules/codegen_console/infrastructure/repo"
	codegen_consolehttp "goadmin/modules/codegen_console/transport/http"
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
	return codegen_consolerepo.Migrate(db)
}

func (Bootstrap) Register(group coretransport.RouteRegistrar, deps corebootstrapcontract.Dependencies) error {
	if group == nil {
		return fmt.Errorf("codegen_console bootstrap requires route registrar")
	}
	if deps.DB == nil {
		return fmt.Errorf("codegen_console bootstrap requires db")
	}
	repo, err := codegen_consolerepo.NewGormRepository(deps.DB)
	if err != nil {
		return err
	}
	service, err := codegen_consoleservice.New(repo)
	if err != nil {
		return err
	}
	codegen_consolehttp.Register(group, codegen_consolehttp.Dependencies{
		Service: service,
		Logger:  deps.Logger,
	})
	return nil
}
