package kratos

import (
	"context"

	coretransport "goadmin/core/transport"
)

type Context struct {
	coretransport.Context
}

type HandlerFunc = coretransport.HandlerFunc

type Middleware = coretransport.Middleware

type RouteRegistrar = coretransport.RouteRegistrar

type Router struct {
	coretransport.Router
}

func NewContext(ctx coretransport.Context) *Context {
	if ctx == nil {
		return &Context{}
	}
	return &Context{Context: ctx}
}

func NewRouter(router coretransport.Router) *Router {
	if router == nil {
		return &Router{}
	}
	return &Router{Router: router}
}

func (c *Context) RequestContext() context.Context {
	if c == nil || c.Context == nil {
		return context.Background()
	}
	return c.Context.RequestContext()
}

func (c *Context) SetRequestContext(ctx context.Context) {
	if c == nil || c.Context == nil {
		return
	}
	c.Context.SetRequestContext(ctx)
}

func (c *Context) Method() string {
	if c == nil || c.Context == nil {
		return ""
	}
	return c.Context.Method()
}

func (c *Context) Path() string {
	if c == nil || c.Context == nil {
		return ""
	}
	return c.Context.Path()
}

func (c *Context) Header(key string) string {
	if c == nil || c.Context == nil {
		return ""
	}
	return c.Context.Header(key)
}

func (c *Context) Param(key string) string {
	if c == nil || c.Context == nil {
		return ""
	}
	return c.Context.Param(key)
}

func (c *Context) Query(key string) string {
	if c == nil || c.Context == nil {
		return ""
	}
	return c.Context.Query(key)
}

func (c *Context) Set(key string, value any) {
	if c == nil || c.Context == nil {
		return
	}
	c.Context.Set(key, value)
}

func (c *Context) Get(key string) (any, bool) {
	if c == nil || c.Context == nil {
		return nil, false
	}
	return c.Context.Get(key)
}

func (c *Context) ShouldBindJSON(v any) error {
	if c == nil || c.Context == nil {
		return context.Canceled
	}
	return c.Context.ShouldBindJSON(v)
}

func (c *Context) ShouldBindQuery(v any) error {
	if c == nil || c.Context == nil {
		return context.Canceled
	}
	return c.Context.ShouldBindQuery(v)
}

func (c *Context) BindJSON(v any) error {
	return c.ShouldBindJSON(v)
}

func (c *Context) JSON(status int, payload any) {
	if c == nil || c.Context == nil {
		return
	}
	c.Context.JSON(status, payload)
}

func (c *Context) AbortWithStatusJSON(status int, payload any) {
	if c == nil || c.Context == nil {
		return
	}
	c.Context.AbortWithStatusJSON(status, payload)
}

func (r *Router) Use(middlewares ...Middleware) {
	if r == nil || r.Router == nil {
		return
	}
	r.Router.Use(middlewares...)
}

func (r *Router) Group(path string, middlewares ...Middleware) RouteRegistrar {
	if r == nil || r.Router == nil {
		return nil
	}
	return r.Router.Group(path, middlewares...)
}

func (r *Router) GET(path string, handler HandlerFunc, middlewares ...Middleware) {
	if r == nil || r.Router == nil {
		return
	}
	r.Router.GET(path, handler, middlewares...)
}

func (r *Router) POST(path string, handler HandlerFunc, middlewares ...Middleware) {
	if r == nil || r.Router == nil {
		return
	}
	r.Router.POST(path, handler, middlewares...)
}

func (r *Router) PUT(path string, handler HandlerFunc, middlewares ...Middleware) {
	if r == nil || r.Router == nil {
		return
	}
	r.Router.PUT(path, handler, middlewares...)
}

func (r *Router) PATCH(path string, handler HandlerFunc, middlewares ...Middleware) {
	if r == nil || r.Router == nil {
		return
	}
	r.Router.PATCH(path, handler, middlewares...)
}

func (r *Router) DELETE(path string, handler HandlerFunc, middlewares ...Middleware) {
	if r == nil || r.Router == nil {
		return
	}
	r.Router.DELETE(path, handler, middlewares...)
}

func (r *Router) Any(path string, handler HandlerFunc, middlewares ...Middleware) {
	if r == nil || r.Router == nil {
		return
	}
	r.Router.Any(path, handler, middlewares...)
}

func AsCoreRegistrar(registrar RouteRegistrar) coretransport.RouteRegistrar {
	return registrar
}
