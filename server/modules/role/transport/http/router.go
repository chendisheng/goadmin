package rolehttp

import (
	coretransport "goadmin/core/transport"
	roleservice "goadmin/modules/role/application/service"
	"goadmin/modules/role/transport/http/handler"

	"go.uber.org/zap"
)

type Dependencies struct {
	Service *roleservice.Service
	Logger  *zap.Logger
}

func Register(group coretransport.RouteRegistrar, deps Dependencies) {
	h := handler.New(deps.Service, deps.Logger)
	root := group.Group("/roles")
	root.GET("", h.List)
	root.GET("/:id", h.Get)
	root.POST("", h.Create)
	root.PUT("/:id", h.Update)
	root.DELETE("/:id", h.Delete)
}
