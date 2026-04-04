package http

import (
	"go.uber.org/zap"
	coretransport "goadmin/core/transport"
	orderservice "goadmin/modules/order/application/service"
	"goadmin/modules/order/transport/http/handler"
)

type Dependencies struct {
	Service *orderservice.Service
	Logger  *zap.Logger
}

func Register(group coretransport.RouteRegistrar, deps Dependencies) {
	h := handler.New(deps.Service, deps.Logger)
	root := group.Group("/orders")
	root.GET("", h.List)
	root.GET("/:id", h.Get)
	root.POST("", h.Create)
	root.PUT("/:id", h.Update)
	root.DELETE("/:id", h.Delete)
}
