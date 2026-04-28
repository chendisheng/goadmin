package http

import (
	coreauthbootstrap "goadmin/core/auth/bootstrap"
	coreauthjwt "goadmin/core/auth/jwt"
	coretransport "goadmin/core/transport"
	pluginservice "goadmin/plugin/application/service"
	"goadmin/plugin/transport/http/handler"
	ginmiddleware "goadmin/transport/http/gin/middleware"

	"go.uber.org/zap"
)

type Dependencies struct {
	Service     *pluginservice.Service
	Logger      *zap.Logger
	JWT         *coreauthjwt.Manager
	Authorizer  coreauthbootstrap.Authorizer
	Revocations coreauthbootstrap.RevocationStore
}

func Register(group coretransport.RouteRegistrar, deps Dependencies) {
	h := handler.New(deps.Service, deps.Logger)
	root := group.Group("/plugins")
	read := root.Group("", ginmiddleware.JWTAuth(deps.JWT, deps.Revocations))
	read.GET("/menus", h.Menus)
	read.GET("/permissions", h.Permissions)

	protected := root.Group("", ginmiddleware.JWTAuth(deps.JWT, deps.Revocations), ginmiddleware.RequirePermission(deps.Authorizer))
	protected.GET("", h.List)
	protected.GET("/:name", h.Get)
	protected.POST("", h.Create)
	protected.PUT("/:name", h.Update)
	protected.DELETE("/:name", h.Delete)
}
