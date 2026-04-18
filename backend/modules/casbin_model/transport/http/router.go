package http

import (
	"go.uber.org/zap"
	coretransport "goadmin/core/transport"
	casbin_modelservice "goadmin/modules/casbin_model/application/service"
	"goadmin/modules/casbin_model/transport/http/handler"
)

type Dependencies struct {
	Service *casbin_modelservice.Service
	Logger  *zap.Logger
}

func Register(group coretransport.RouteRegistrar, deps Dependencies) {
	h := handler.New(deps.Service, deps.Logger)
	root := group.Group("/casbin_models")
	root.GET("", h.List)
	root.GET("/:id", h.Get)
	root.POST("", h.Create)
	root.PUT("/:id", h.Update)
	root.DELETE("/:id", h.Delete)
}
