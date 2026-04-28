package casbin

import (
	corebootstrapcontract "goadmin/core/bootstrap/contract"
	apperrors "goadmin/core/errors"
	coretransport "goadmin/core/transport"
	casbinservice "goadmin/modules/casbin/application/service"
	casbinhttp "goadmin/modules/casbin/transport/http"
	casbinmodel "goadmin/modules/casbin_model"
	casbinrule "goadmin/modules/casbin_rule"

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
	if err := casbinmodel.NewBootstrap().Migrate(db); err != nil {
		return err
	}
	return casbinrule.NewBootstrap().Migrate(db)
}

func (Bootstrap) Register(group coretransport.RouteRegistrar, deps corebootstrapcontract.Dependencies) error {
	if group == nil {
		return apperrors.NewWithKey(apperrors.CodeInternal, "casbin.bootstrap_route_registrar_required", "casbin bootstrap requires route registrar")
	}
	if deps.DB == nil {
		return apperrors.NewWithKey(apperrors.CodeInternal, "casbin.bootstrap_db_required", "casbin bootstrap requires db")
	}

	service, err := casbinservice.New(casbinservice.Config{
		Config:               deps.Config,
		AuthorizationRuntime: deps.AuthorizationRuntime,
	})
	if err != nil {
		return err
	}

	casbinhttp.Register(group, casbinhttp.Dependencies{Service: service, Logger: deps.Logger})
	return nil
}
