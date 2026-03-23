package menuhttp

import (
	coretransport "goadmin/core/transport"
	menuservice "goadmin/modules/menu/application/service"
	"goadmin/modules/menu/transport/http/handler"

	"go.uber.org/zap"
)

type Dependencies struct {
	Service *menuservice.Service
	Logger  *zap.Logger
}

func Register(group coretransport.RouteRegistrar, deps Dependencies) {
	h := handler.New(deps.Service, deps.Logger)
	root := group.Group("/menus")
	root.GET("", h.List)
	root.GET("/:id", h.Get)
	root.GET("/tree", h.Tree)
	root.GET("/routes", h.Routes)
	root.POST("", h.Create)
	root.PUT("/:id", h.Update)
	root.DELETE("/:id", h.Delete)
}
