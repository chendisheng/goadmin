package http

import (
	"go.uber.org/zap"
	coretransport "goadmin/core/transport"
	bookservice "goadmin/modules/book/application/service"
	"goadmin/modules/book/transport/http/handler"
)

type Dependencies struct {
	Service *bookservice.Service
	Logger  *zap.Logger
}

func Register(group coretransport.RouteRegistrar, deps Dependencies) {
	h := handler.New(deps.Service, deps.Logger)
	root := group.Group("/books")
	root.GET("", h.List)
	root.GET("/:id", h.Get)
	root.POST("", h.Create)
	root.PUT("/:id", h.Update)
	root.DELETE("/:id", h.Delete)
}
