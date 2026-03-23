package http

import (
	coretransport "goadmin/core/transport"
	pluginservice "goadmin/plugin/application/service"
	"goadmin/plugin/transport/http/handler"

	"go.uber.org/zap"
)

type Dependencies struct {
	Service *pluginservice.Service
	Logger  *zap.Logger
}

func Register(group coretransport.RouteRegistrar, deps Dependencies) {
	h := handler.New(deps.Service, deps.Logger)
	root := group.Group("/plugins")
	root.GET("", h.List)
	root.GET("/:name", h.Get)
	root.POST("", h.Create)
	root.PUT("/:name", h.Update)
	root.DELETE("/:name", h.Delete)
	root.GET("/menus", h.Menus)
	root.GET("/permissions", h.Permissions)
}
