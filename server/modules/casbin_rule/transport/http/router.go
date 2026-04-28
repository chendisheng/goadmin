package http

import (
	"go.uber.org/zap"
	coretransport "goadmin/core/transport"
	casbin_ruleservice "goadmin/modules/casbin_rule/application/service"
	"goadmin/modules/casbin_rule/transport/http/handler"
)

type Dependencies struct {
	Service *casbin_ruleservice.Service
	Logger  *zap.Logger
}

func Register(group coretransport.RouteRegistrar, deps Dependencies) {
	h := handler.New(deps.Service, deps.Logger)
	root := group.Group("/casbin_rules")
	root.GET("", h.List)
	root.GET("/:id", h.Get)
	root.POST("", h.Create)
	root.PUT("/:id", h.Update)
	root.DELETE("/:id", h.Delete)
}
