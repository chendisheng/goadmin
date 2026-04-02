package router

import (
	"fmt"
	"net/http"
	"time"

	codegenhttp "goadmin/codegen/transport/http"
	coreauthbootstrap "goadmin/core/auth/bootstrap"
	coreauthjwt "goadmin/core/auth/jwt"
	"goadmin/core/config"
	apperrors "goadmin/core/errors"
	coremiddleware "goadmin/core/middleware"
	"goadmin/core/response"
	coretransport "goadmin/core/transport"
	authservice "goadmin/modules/auth/application/service"
	authhttp "goadmin/modules/auth/transport/http"
	menuservice "goadmin/modules/menu/application/service"
	menuhttp "goadmin/modules/menu/transport/http"
	roleservice "goadmin/modules/role/application/service"
	rolehttp "goadmin/modules/role/transport/http"
	userservice "goadmin/modules/user/application/service"
	userhttp "goadmin/modules/user/transport/http"
	pluginservice "goadmin/plugin/application/service"
	pluginiface "goadmin/plugin/interface"
	pluginregistry "goadmin/plugin/registry"
	pluginhttp "goadmin/plugin/transport/http"
	"goadmin/transport/http/gin/handler"
	ginmiddleware "goadmin/transport/http/gin/middleware"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Dependencies struct {
	Config         *config.Config
	Logger         *zap.Logger
	Started        time.Time
	AuthService    *authservice.Service
	UserService    *userservice.Service
	RoleService    *roleservice.Service
	MenuService    *menuservice.Service
	PluginService  *pluginservice.Service
	PluginRegistry *pluginregistry.Registry
	ProjectRoot    string
	JWT            *coreauthjwt.Manager
	Authorizer     coreauthbootstrap.Authorizer
	Revocations    coreauthbootstrap.RevocationStore
}

func Register(engine *gin.Engine, deps Dependencies) {
	h := handler.NewHealthHandler(deps.Config, deps.Logger, deps.Started)

	engine.Use(ginmiddleware.RequestID())
	engine.Use(ginmiddleware.AccessLog(deps.Logger))
	engine.Use(ginmiddleware.Recovery(deps.Logger))
	engine.Use(ginmiddleware.CORS())

	api := newRouteRegistrarAdapter(engine.Group("/api/v1"))
	{
		api.GET("/health", h.Health)
		api.GET("/meta/version", h.Version)
		api.GET("/meta/config", h.Config)

		authhttp.Register(api, authhttp.Dependencies{
			Service:     deps.AuthService,
			Logger:      deps.Logger,
			JWT:         deps.JWT,
			Authorizer:  deps.Authorizer,
			Revocations: deps.Revocations,
		})

		protected := api.Group("", ginmiddleware.JWTAuth(deps.JWT, deps.Revocations), ginmiddleware.RequirePermission(deps.Authorizer))

		userhttp.Register(protected, userhttp.Dependencies{Service: deps.UserService, Logger: deps.Logger})
		rolehttp.Register(protected, rolehttp.Dependencies{Service: deps.RoleService, Logger: deps.Logger})
		menuhttp.Register(protected, menuhttp.Dependencies{Service: deps.MenuService, Logger: deps.Logger})
		codegenhttp.Register(protected, codegenhttp.Dependencies{ProjectRoot: deps.ProjectRoot})
		pluginhttp.Register(protected, pluginhttp.Dependencies{Service: deps.PluginService, Logger: deps.Logger})
		registerPluginRoutes(api, protected, deps.PluginRegistry)
	}

	engine.NoRoute(func(c *gin.Context) {
		status, body := response.Failure(apperrors.New(apperrors.CodeNotFound, fmt.Sprintf("route %s %s not found", c.Request.Method, c.Request.URL.Path)), requestID(c))
		c.JSON(status, body)
	})
}

func requestID(c *gin.Context) string {
	if value, exists := c.Get(coremiddleware.RequestIDContextKey); exists {
		if id, ok := value.(string); ok {
			return id
		}
	}
	return ""
}

func registerPluginRoutes(public, protected coretransport.RouteRegistrar, registry *pluginregistry.Registry) {
	if registry == nil {
		return
	}
	for _, route := range registry.Routes() {
		target := protected
		if route.Access == pluginiface.AccessPublic {
			target = public
		}
		if target == nil {
			continue
		}
		registerPluginRoute(target, route)
	}
}

func registerPluginRoute(group coretransport.RouteRegistrar, route pluginiface.Route) {
	if group == nil || route.Handler == nil {
		return
	}
	switch route.Method {
	case http.MethodGet:
		group.GET(route.Path, route.Handler, route.Middlewares...)
	case http.MethodPost:
		group.POST(route.Path, route.Handler, route.Middlewares...)
	case http.MethodPut:
		group.PUT(route.Path, route.Handler, route.Middlewares...)
	case http.MethodPatch:
		group.PATCH(route.Path, route.Handler, route.Middlewares...)
	case http.MethodDelete:
		group.DELETE(route.Path, route.Handler, route.Middlewares...)
	default:
		group.Any(route.Path, route.Handler, route.Middlewares...)
	}
}
