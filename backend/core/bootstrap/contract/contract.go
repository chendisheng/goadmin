package contract

import (
	"goadmin/core/config"
	"goadmin/core/event"
	coretransport "goadmin/core/transport"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type AuthorizationRuntime interface {
	Reload() error
	SeedDefaultPolicy() error
	String() string
}

type CasbinRuntime = AuthorizationRuntime

type Dependencies struct {
	DB                   *gorm.DB
	Logger               *zap.Logger
	EventBus             event.Bus
	Config               *config.Config
	AuthorizationRuntime AuthorizationRuntime
}

type Module interface {
	Name() string
	ManifestPath() string
	Migrate(db *gorm.DB) error
	Register(group coretransport.RouteRegistrar, deps Dependencies) error
}

func MigrateAll(db *gorm.DB, modules []Module) error {
	for _, module := range modules {
		if module == nil {
			continue
		}
		if err := module.Migrate(db); err != nil {
			return err
		}
	}
	return nil
}

func RegisterAll(group coretransport.RouteRegistrar, deps Dependencies, modules []Module) error {
	for _, module := range modules {
		if module == nil {
			continue
		}
		if err := module.Register(group, deps); err != nil {
			return err
		}
	}
	return nil
}
