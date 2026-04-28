package dictionary

import (
	"fmt"

	corebootstrapcontract "goadmin/core/bootstrap/contract"
	coretransport "goadmin/core/transport"
	dictionaryservice "goadmin/modules/dictionary/application/service"
	dictionaryrepo "goadmin/modules/dictionary/infrastructure/repo"
	dictionaryhttp "goadmin/modules/dictionary/transport/http"
	dictionaryhandler "goadmin/modules/dictionary/transport/http/handler"

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
	return dictionaryrepo.Migrate(db)
}

func (Bootstrap) Register(group coretransport.RouteRegistrar, deps corebootstrapcontract.Dependencies) error {
	if group == nil {
		return fmt.Errorf("dictionary bootstrap requires route registrar")
	}
	if deps.DB == nil {
		return fmt.Errorf("dictionary bootstrap requires db")
	}
	categoryRepo, err := dictionaryrepo.NewCategoryRepository(deps.DB)
	if err != nil {
		return err
	}
	itemRepo, err := dictionaryrepo.NewItemRepository(deps.DB)
	if err != nil {
		return err
	}
	service, err := dictionaryservice.New(categoryRepo, itemRepo)
	if err != nil {
		return err
	}
	handler := dictionaryhandler.New(service, deps.Logger)
	dictionaryhttp.Register(group, dictionaryhttp.Dependencies{Handler: handler, Logger: deps.Logger})
	return nil
}
