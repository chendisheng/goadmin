package http

import (
	"go.uber.org/zap"
	coretransport "goadmin/core/transport"
	casbinservice "goadmin/modules/casbin/application/service"
	"goadmin/modules/casbin/transport/http/handler"
)

type Dependencies struct {
	Service *casbinservice.Service
	Logger  *zap.Logger
}

func Register(group coretransport.RouteRegistrar, deps Dependencies) {
	h := handler.New(deps.Service, deps.Logger)
	root := group.Group("/casbin")
	root.GET("/status", h.Status)
	root.POST("/reload", h.Reload)
	root.POST("/seed", h.Seed)
}
