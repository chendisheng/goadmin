package userhttp

import (
	coretransport "goadmin/core/transport"
	userservice "goadmin/modules/user/application/service"
	"goadmin/modules/user/transport/http/handler"

	"go.uber.org/zap"
)

type Dependencies struct {
	Service *userservice.Service
	Logger  *zap.Logger
}

func Register(group coretransport.RouteRegistrar, deps Dependencies) {
	h := handler.New(deps.Service, deps.Logger)
	root := group.Group("/users")
	root.GET("", h.List)
	root.GET("/:id", h.Get)
	root.POST("", h.Create)
	root.PUT("/:id", h.Update)
	root.DELETE("/:id", h.Delete)
}
