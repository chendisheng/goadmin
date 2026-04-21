package http

import (
	coretransport "goadmin/core/transport"
	uploadservice "goadmin/modules/upload/application/service"
	"goadmin/modules/upload/transport/http/handler"

	"go.uber.org/zap"
)

type Dependencies struct {
	Service *uploadservice.Service
	Logger  *zap.Logger
}

func Register(group coretransport.RouteRegistrar, deps Dependencies) {
	h := handler.New(deps.Service, deps.Logger)
	root := group.Group("/uploads/files")

	root.GET("/storage/default", h.GetDefaultStorage)
	root.PUT("/storage/default", h.SetDefaultStorage)
	root.GET("", h.List)
	root.GET("/:id", h.Get)
	root.POST("", h.Upload)
	root.DELETE("/:id", h.Delete)
	root.GET("/:id/download", h.Download)
	root.GET("/:id/preview", h.Preview)
	root.POST("/:id/bind", h.Bind)
	root.DELETE("/:id/bind", h.Unbind)
}
