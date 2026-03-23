package kratos

import (
	"context"
	"testing"

	coretransport "goadmin/core/transport"
)

type fakeContext struct {
	requestContext  context.Context
	method          string
	path            string
	headers         map[string]string
	params          map[string]string
	queries         map[string]string
	values          map[string]any
	bindQueryCalled bool
}

func (c *fakeContext) RequestContext() context.Context {
	if c.requestContext == nil {
		return context.Background()
	}
	return c.requestContext
}

func (c *fakeContext) SetRequestContext(ctx context.Context) { c.requestContext = ctx }

func (c *fakeContext) Method() string           { return c.method }
func (c *fakeContext) Path() string             { return c.path }
func (c *fakeContext) Header(key string) string { return c.headers[key] }
func (c *fakeContext) Param(key string) string  { return c.params[key] }
func (c *fakeContext) Query(key string) string  { return c.queries[key] }
func (c *fakeContext) Set(key string, value any) {
	if c.values == nil {
		c.values = make(map[string]any)
	}
	c.values[key] = value
}
func (c *fakeContext) Get(key string) (any, bool) {
	value, ok := c.values[key]
	return value, ok
}
func (c *fakeContext) ShouldBindJSON(any) error     { return nil }
func (c *fakeContext) ShouldBindQuery(any) error    { c.bindQueryCalled = true; return nil }
func (c *fakeContext) BindJSON(v any) error         { return c.ShouldBindJSON(v) }
func (c *fakeContext) JSON(int, any)                {}
func (c *fakeContext) AbortWithStatusJSON(int, any) {}

type fakeRouter struct {
	used   int
	groups []string
	routes []string
}

func (r *fakeRouter) Use(middlewares ...coretransport.Middleware) { r.used += len(middlewares) }
func (r *fakeRouter) Group(path string, middlewares ...coretransport.Middleware) coretransport.RouteRegistrar {
	r.groups = append(r.groups, path)
	r.used += len(middlewares)
	return r
}
func (r *fakeRouter) GET(path string, _ coretransport.HandlerFunc, middlewares ...coretransport.Middleware) {
	r.routes = append(r.routes, "GET "+path)
	r.used += len(middlewares)
}
func (r *fakeRouter) POST(path string, _ coretransport.HandlerFunc, middlewares ...coretransport.Middleware) {
	r.routes = append(r.routes, "POST "+path)
	r.used += len(middlewares)
}
func (r *fakeRouter) PUT(path string, _ coretransport.HandlerFunc, middlewares ...coretransport.Middleware) {
	r.routes = append(r.routes, "PUT "+path)
	r.used += len(middlewares)
}
func (r *fakeRouter) PATCH(path string, _ coretransport.HandlerFunc, middlewares ...coretransport.Middleware) {
	r.routes = append(r.routes, "PATCH "+path)
	r.used += len(middlewares)
}
func (r *fakeRouter) DELETE(path string, _ coretransport.HandlerFunc, middlewares ...coretransport.Middleware) {
	r.routes = append(r.routes, "DELETE "+path)
	r.used += len(middlewares)
}
func (r *fakeRouter) Any(path string, _ coretransport.HandlerFunc, middlewares ...coretransport.Middleware) {
	r.routes = append(r.routes, "ANY "+path)
	r.used += len(middlewares)
}

func TestContextWrapperDelegates(t *testing.T) {
	t.Parallel()

	base := &fakeContext{method: "POST", path: "/items", headers: map[string]string{"X": "2"}}
	ctx := NewContext(base)
	if got := ctx.Method(); got != "POST" {
		t.Fatalf("Method() = %q, want POST", got)
	}
	if got := ctx.Path(); got != "/items" {
		t.Fatalf("Path() = %q, want /items", got)
	}
	if got := ctx.Header("X"); got != "2" {
		t.Fatalf("Header(X) = %q, want 2", got)
	}
	if err := ctx.ShouldBindQuery(nil); err != nil {
		t.Fatalf("ShouldBindQuery returned error: %v", err)
	}
	if !base.bindQueryCalled {
		t.Fatal("expected ShouldBindQuery to delegate")
	}
}

func TestRouterWrapperDelegates(t *testing.T) {
	t.Parallel()

	base := &fakeRouter{}
	router := NewRouter(base)
	router.Use(nil)
	group := router.Group("/v1")
	group.POST("/echo", func(coretransport.Context) {})

	if len(base.groups) != 1 || base.groups[0] != "/v1" {
		t.Fatalf("groups = %v, want [/v1]", base.groups)
	}
	if len(base.routes) != 1 || base.routes[0] != "POST /echo" {
		t.Fatalf("routes = %v, want [POST /echo]", base.routes)
	}
}
