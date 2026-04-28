package upload

import (
	"context"

	corebootstrapcontract "goadmin/core/bootstrap/contract"
	apperrors "goadmin/core/errors"
	coretransport "goadmin/core/transport"
	uploadservice "goadmin/modules/upload/application/service"
	uploadpersist "goadmin/modules/upload/infrastructure/persistence/gorm"
	uploadstorage "goadmin/modules/upload/infrastructure/storage"
	uploadhttp "goadmin/modules/upload/transport/http"

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
	return uploadpersist.Migrate(db)
}

func (Bootstrap) Register(group coretransport.RouteRegistrar, deps corebootstrapcontract.Dependencies) error {
	if group == nil {
		return apperrors.NewWithKey(apperrors.CodeInternal, "upload.bootstrap_route_registrar_required", "upload bootstrap requires route registrar")
	}
	if deps.DB == nil {
		return apperrors.NewWithKey(apperrors.CodeInternal, "upload.bootstrap_db_required", "upload bootstrap requires db")
	}
	if deps.Config == nil {
		return apperrors.NewWithKey(apperrors.CodeInternal, "upload.bootstrap_config_required", "upload bootstrap requires config")
	}
	repo, err := uploadpersist.New(deps.DB)
	if err != nil {
		return err
	}
	storageCfg := deps.Config.Upload.Storage
	defaultDriver, err := repo.DefaultStorageDriver(context.Background(), storageCfg.Driver)
	if err != nil {
		return err
	}
	storageCfg.Driver = defaultDriver
	driver, err := uploadstorage.NewDriverWithDB(deps.DB, storageCfg)
	if err != nil {
		return err
	}
	service, err := uploadservice.New(repo, driver, deps.Config.Upload.Storage.Policy)
	if err != nil {
		return err
	}
	uploadhttp.Register(group, uploadhttp.Dependencies{
		Service: service,
		Logger:  deps.Logger,
	})
	return nil
}
