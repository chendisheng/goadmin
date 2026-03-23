package transport_test

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	coretransport "goadmin/core/transport"
	authhttp "goadmin/modules/auth/transport/http"
	menuhttp "goadmin/modules/menu/transport/http"
	rolehttp "goadmin/modules/role/transport/http"
	userhttp "goadmin/modules/user/transport/http"
	pluginhttp "goadmin/plugin/transport/http"
)

type recordedRoute struct {
	method      string
	path        string
	middlewares int
}

type fakeRegistrar struct {
	prefix      string
	middlewares int
	routes      *[]recordedRoute
}

func newFakeRegistrar(routes *[]recordedRoute) coretransport.RouteRegistrar {
	return &fakeRegistrar{routes: routes}
}

func (r *fakeRegistrar) Group(path string, middlewares ...coretransport.Middleware) coretransport.RouteRegistrar {
	if r == nil {
		return &fakeRegistrar{routes: r.routes}
	}
	return &fakeRegistrar{
		prefix:      joinPath(r.prefix, path),
		middlewares: r.middlewares + len(middlewares),
		routes:      r.routes,
	}
}

func (r *fakeRegistrar) GET(path string, _ coretransport.HandlerFunc, middlewares ...coretransport.Middleware) {
	r.record("GET", path, middlewares)
}

func (r *fakeRegistrar) POST(path string, _ coretransport.HandlerFunc, middlewares ...coretransport.Middleware) {
	r.record("POST", path, middlewares)
}

func (r *fakeRegistrar) PUT(path string, _ coretransport.HandlerFunc, middlewares ...coretransport.Middleware) {
	r.record("PUT", path, middlewares)
}

func (r *fakeRegistrar) PATCH(path string, _ coretransport.HandlerFunc, middlewares ...coretransport.Middleware) {
	r.record("PATCH", path, middlewares)
}

func (r *fakeRegistrar) DELETE(path string, _ coretransport.HandlerFunc, middlewares ...coretransport.Middleware) {
	r.record("DELETE", path, middlewares)
}

func (r *fakeRegistrar) Any(path string, _ coretransport.HandlerFunc, middlewares ...coretransport.Middleware) {
	r.record("ANY", path, middlewares)
}

func (r *fakeRegistrar) record(method, path string, middlewares []coretransport.Middleware) {
	if r == nil || r.routes == nil {
		return
	}
	fullPath := joinPath(r.prefix, path)
	*r.routes = append(*r.routes, recordedRoute{
		method:      method,
		path:        fullPath,
		middlewares: r.middlewares + len(middlewares),
	})
}

func joinPath(prefix, path string) string {
	prefix = strings.TrimRight(strings.TrimSpace(prefix), "/")
	path = strings.TrimSpace(path)
	if prefix == "" {
		if path == "" {
			return "/"
		}
		if strings.HasPrefix(path, "/") {
			return path
		}
		return "/" + path
	}
	if path == "" {
		return prefix
	}
	if strings.HasPrefix(path, "/") {
		return prefix + path
	}
	return prefix + "/" + path
}

func TestModuleRoutesCanRegisterThroughCoreTransport(t *testing.T) {
	t.Parallel()

	var routes []recordedRoute
	root := newFakeRegistrar(&routes)

	authhttp.Register(root, authhttp.Dependencies{})
	userhttp.Register(root, userhttp.Dependencies{})
	rolehttp.Register(root, rolehttp.Dependencies{})
	menuhttp.Register(root, menuhttp.Dependencies{})
	pluginhttp.Register(root, pluginhttp.Dependencies{})

	requireRoutes(t, routes, []recordedRoute{
		{method: "POST", path: "/auth/login"},
		{method: "POST", path: "/auth/logout"},
		{method: "GET", path: "/auth/me"},
		{method: "GET", path: "/users"},
		{method: "GET", path: "/roles"},
		{method: "GET", path: "/menus"},
		{method: "GET", path: "/plugins"},
	})

	for _, route := range routes {
		if route.path == "/auth/logout" && route.middlewares == 0 {
			t.Fatalf("expected auth protected route to carry middlewares, got %+v", route)
		}
	}
}

func requireRoutes(t *testing.T, got []recordedRoute, want []recordedRoute) {
	t.Helper()

	index := make(map[string]recordedRoute, len(got))
	for _, route := range got {
		index[fmt.Sprintf("%s %s", route.method, route.path)] = route
	}
	for _, route := range want {
		key := fmt.Sprintf("%s %s", route.method, route.path)
		if _, ok := index[key]; !ok {
			t.Fatalf("missing route %s; got=%v", key, got)
		}
	}
	if len(got) == 0 {
		t.Fatal("expected routes to be recorded")
	}
	paths := make([]string, 0, len(got))
	for _, route := range got {
		paths = append(paths, fmt.Sprintf("%s %s", route.method, route.path))
	}
	sort.Strings(paths)
	_ = paths
}
