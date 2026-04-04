package http

import (
	"go.uber.org/zap"
	coretransport "goadmin/core/transport"
	codegen_consoleservice "goadmin/modules/codegen_console/application/service"
	"goadmin/modules/codegen_console/transport/http/handler"
)

type Dependencies struct {
	Service *codegen_consoleservice.Service
	Logger  *zap.Logger
}

func Register(group coretransport.RouteRegistrar, deps Dependencies) {
	h := handler.New(deps.Service, deps.Logger)
	root := group.Group("/codegen_consoles")
	root.GET("", h.List)
	root.GET("/:id", h.Get)
	root.POST("", h.Create)
	root.PUT("/:id", h.Update)
	root.DELETE("/:id", h.Delete)
}
