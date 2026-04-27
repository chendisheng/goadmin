package pluginiface

import (
	coreconfig "goadmin/core/config"
	coreregistry "goadmin/core/registry"
	coretransport "goadmin/core/transport"

	"go.uber.org/zap"
)

type Access string

const (
	AccessPublic    Access = "public"
	AccessProtected Access = "protected"
)

type MenuType string

const (
	MenuTypeDirectory MenuType = "directory"
	MenuTypeMenu      MenuType = "menu"
	MenuTypeButton    MenuType = "button"
)

type Route struct {
	Plugin      string
	Name        string
	Method      string
	Path        string
	Access      Access
	Handler     coretransport.HandlerFunc
	Middlewares []coretransport.Middleware
}

type Menu struct {
	Plugin       string
	ID           string
	ParentID     string
	Name         string
	TitleKey     string
	TitleDefault string
	Path         string
	Component    string
	Icon         string
	Sort         int
	Permission   string
	Type         MenuType
	Visible      bool
	Enabled      bool
	Redirect     string
	ExternalURL  string
	Children     []Menu
}

type Permission struct {
	Plugin      string
	Object      string
	Action      string
	Description string
}

type Context struct {
	Config    *coreconfig.Config
	Logger    *zap.Logger
	Container *coreregistry.Container
}

type Registrar interface {
	RegisterPlugin(name string) error
	AddRoute(Route) error
	AddMenu(Menu) error
	AddPermission(Permission) error
}

type Plugin interface {
	Name() string
	Register(*Context, Registrar) error
}
