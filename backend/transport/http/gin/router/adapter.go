package router

import (
	"context"

	coretransport "goadmin/core/transport"

	"github.com/gin-gonic/gin"
)

type routeRegistrarAdapter struct {
	group       *gin.RouterGroup
	middlewares []coretransport.Middleware
}

func newRouteRegistrarAdapter(group *gin.RouterGroup) coretransport.RouteRegistrar {
	return &routeRegistrarAdapter{group: group}
}

func (r *routeRegistrarAdapter) Group(path string, middlewares ...coretransport.Middleware) coretransport.RouteRegistrar {
	if r == nil || r.group == nil {
		return &routeRegistrarAdapter{middlewares: append([]coretransport.Middleware(nil), middlewares...)}
	}
	child := r.group.Group(path)
	chain := append([]coretransport.Middleware(nil), r.middlewares...)
	chain = append(chain, middlewares...)
	return &routeRegistrarAdapter{group: child, middlewares: chain}
}

func (r *routeRegistrarAdapter) GET(path string, handler coretransport.HandlerFunc, middlewares ...coretransport.Middleware) {
	r.register(path, handler, middlewares, "GET")
}

func (r *routeRegistrarAdapter) POST(path string, handler coretransport.HandlerFunc, middlewares ...coretransport.Middleware) {
	r.register(path, handler, middlewares, "POST")
}

func (r *routeRegistrarAdapter) PUT(path string, handler coretransport.HandlerFunc, middlewares ...coretransport.Middleware) {
	r.register(path, handler, middlewares, "PUT")
}

func (r *routeRegistrarAdapter) PATCH(path string, handler coretransport.HandlerFunc, middlewares ...coretransport.Middleware) {
	r.register(path, handler, middlewares, "PATCH")
}

func (r *routeRegistrarAdapter) DELETE(path string, handler coretransport.HandlerFunc, middlewares ...coretransport.Middleware) {
	r.register(path, handler, middlewares, "DELETE")
}

func (r *routeRegistrarAdapter) Any(path string, handler coretransport.HandlerFunc, middlewares ...coretransport.Middleware) {
	r.register(path, handler, middlewares, "ANY")
}

func (r *routeRegistrarAdapter) register(path string, handler coretransport.HandlerFunc, middlewares []coretransport.Middleware, method string) {
	if r == nil || r.group == nil || handler == nil {
		return
	}
	wrapped := adaptCoreHandler(handler, append(r.middlewares, middlewares...))
	switch method {
	case "GET":
		r.group.GET(path, wrapped)
	case "POST":
		r.group.POST(path, wrapped)
	case "PUT":
		r.group.PUT(path, wrapped)
	case "PATCH":
		r.group.PATCH(path, wrapped)
	case "DELETE":
		r.group.DELETE(path, wrapped)
	default:
		r.group.Any(path, wrapped)
	}
}

func adaptCoreHandler(handler coretransport.HandlerFunc, middlewares []coretransport.Middleware) gin.HandlerFunc {
	chain := handler
	for i := len(middlewares) - 1; i >= 0; i-- {
		if middlewares[i] == nil {
			continue
		}
		chain = middlewares[i](chain)
	}
	return func(c *gin.Context) {
		if chain == nil {
			return
		}
		chain(&ginContextAdapter{Context: c})
	}
}

type ginContextAdapter struct {
	Context *gin.Context
}

func (c *ginContextAdapter) RequestContext() context.Context {
	if c == nil || c.Context == nil || c.Context.Request == nil {
		return context.Background()
	}
	return c.Context.Request.Context()
}

func (c *ginContextAdapter) SetRequestContext(ctx context.Context) {
	if c == nil || c.Context == nil {
		return
	}
	if c.Context.Request == nil {
		return
	}
	c.Context.Request = c.Context.Request.WithContext(ctx)
}

func (c *ginContextAdapter) Method() string {
	if c == nil || c.Context == nil || c.Context.Request == nil {
		return ""
	}
	return c.Context.Request.Method
}

func (c *ginContextAdapter) Path() string {
	if c == nil || c.Context == nil {
		return ""
	}
	if path := c.Context.FullPath(); path != "" {
		return path
	}
	if c.Context.Request != nil {
		return c.Context.Request.URL.Path
	}
	return ""
}

func (c *ginContextAdapter) Header(key string) string {
	if c == nil || c.Context == nil {
		return ""
	}
	return c.Context.GetHeader(key)
}

func (c *ginContextAdapter) SetHeader(key, value string) {
	if c == nil || c.Context == nil {
		return
	}
	c.Context.Header(key, value)
}

func (c *ginContextAdapter) Param(key string) string {
	if c == nil || c.Context == nil {
		return ""
	}
	return c.Context.Param(key)
}

func (c *ginContextAdapter) Query(key string) string {
	if c == nil || c.Context == nil {
		return ""
	}
	return c.Context.Query(key)
}

func (c *ginContextAdapter) Set(key string, value any) {
	if c == nil || c.Context == nil {
		return
	}
	c.Context.Set(key, value)
}

func (c *ginContextAdapter) Get(key string) (any, bool) {
	if c == nil || c.Context == nil {
		return nil, false
	}
	return c.Context.Get(key)
}

func (c *ginContextAdapter) ShouldBindJSON(v any) error {
	if c == nil || c.Context == nil {
		return context.Canceled
	}
	return c.Context.ShouldBindJSON(v)
}

func (c *ginContextAdapter) ShouldBindQuery(v any) error {
	if c == nil || c.Context == nil {
		return context.Canceled
	}
	return c.Context.ShouldBindQuery(v)
}

func (c *ginContextAdapter) BindJSON(v any) error {
	return c.ShouldBindJSON(v)
}

func (c *ginContextAdapter) JSON(status int, payload any) {
	if c == nil || c.Context == nil {
		return
	}
	c.Context.JSON(status, payload)
}

func (c *ginContextAdapter) FileAttachment(path, name string) {
	if c == nil || c.Context == nil {
		return
	}
	c.Context.FileAttachment(path, name)
}

func (c *ginContextAdapter) AbortWithStatusJSON(status int, payload any) {
	if c == nil || c.Context == nil {
		return
	}
	c.Context.AbortWithStatusJSON(status, payload)
}
