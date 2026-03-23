package authhttp

import (
	coreauthbootstrap "goadmin/core/auth/bootstrap"
	coreauthjwt "goadmin/core/auth/jwt"
	coretransport "goadmin/core/transport"
	ginmiddleware "goadmin/transport/http/gin/middleware"

	"go.uber.org/zap"

	"goadmin/modules/auth/application/service"
	handler "goadmin/modules/auth/transport/http/handler"
)

type Dependencies struct {
	Service     *service.Service
	Logger      *zap.Logger
	JWT         *coreauthjwt.Manager
	Authorizer  coreauthbootstrap.Authorizer
	Revocations coreauthbootstrap.RevocationStore
}

func Register(group coretransport.RouteRegistrar, deps Dependencies) {
	h := handler.New(deps.Service, deps.Logger)
	root := group.Group("/auth")
	root.POST("/login", h.Login)

	protected := root.Group("", ginmiddleware.JWTAuth(deps.JWT, deps.Revocations), ginmiddleware.RequirePermission(deps.Authorizer))
	protected.POST("/logout", h.Logout)
	protected.GET("/me", h.Me)
}
