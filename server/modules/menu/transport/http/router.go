package menuhttp

import (
	coreauthbootstrap "goadmin/core/auth/bootstrap"
	coreauthjwt "goadmin/core/auth/jwt"
	coretransport "goadmin/core/transport"
	menuservice "goadmin/modules/menu/application/service"
	"goadmin/modules/menu/transport/http/handler"
	ginmiddleware "goadmin/transport/http/gin/middleware"

	"go.uber.org/zap"
)

type Dependencies struct {
	Service     *menuservice.Service
	Logger      *zap.Logger
	JWT         *coreauthjwt.Manager
	Authorizer  coreauthbootstrap.Authorizer
	Revocations coreauthbootstrap.RevocationStore
}

func Register(group coretransport.RouteRegistrar, deps Dependencies) {
	h := handler.New(deps.Service, deps.Logger)
	root := group.Group("/menus")

	read := root.Group("", ginmiddleware.JWTAuth(deps.JWT, deps.Revocations))
	read.GET("/routes", h.Routes)

	protected := root.Group("", ginmiddleware.JWTAuth(deps.JWT, deps.Revocations), ginmiddleware.RequirePermission(deps.Authorizer))
	protected.GET("", h.List)
	protected.GET("/:id", h.Get)
	protected.GET("/tree", h.Tree)
	protected.POST("", h.Create)
	protected.PUT("/:id", h.Update)
	protected.DELETE("/:id", h.Delete)
}
