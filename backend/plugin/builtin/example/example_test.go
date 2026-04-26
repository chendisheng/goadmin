package exampleplugin

import (
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"testing"

	corei18n "goadmin/core/i18n"
	pluginiface "goadmin/plugin/interface"
)

type recordingRegistrar struct {
	routes      []pluginiface.Route
	menus       []pluginiface.Menu
	permissions []pluginiface.Permission
}

func (r *recordingRegistrar) RegisterPlugin(name string) error { return nil }
func (r *recordingRegistrar) AddRoute(route pluginiface.Route) error {
	r.routes = append(r.routes, route)
	return nil
}
func (r *recordingRegistrar) AddMenu(menu pluginiface.Menu) error {
	r.menus = append(r.menus, menu)
	return nil
}
func (r *recordingRegistrar) AddPermission(permission pluginiface.Permission) error {
	r.permissions = append(r.permissions, permission)
	return nil
}

type recordingContext struct {
	requestContext context.Context
	values         map[string]any
	responseStatus int
	responseBody   map[string]any
}

func (c *recordingContext) RequestContext() context.Context {
	if c == nil || c.requestContext == nil {
		return context.Background()
	}
	return c.requestContext
}

func (c *recordingContext) SetRequestContext(ctx context.Context) {
	if c == nil {
		return
	}
	c.requestContext = ctx
}

func (c *recordingContext) Method() string           { return http.MethodGet }
func (c *recordingContext) Path() string             { return "/plugins/example/ping" }
func (c *recordingContext) Header(string) string     { return "" }
func (c *recordingContext) SetHeader(string, string) {}
func (c *recordingContext) Param(string) string      { return "" }
func (c *recordingContext) Query(string) string      { return "" }
func (c *recordingContext) Set(key string, value any) {
	if c.values == nil {
		c.values = make(map[string]any)
	}
	c.values[key] = value
}
func (c *recordingContext) Get(key string) (any, bool) {
	if c == nil || c.values == nil {
		return nil, false
	}
	value, ok := c.values[key]
	return value, ok
}
func (c *recordingContext) ShouldBind(any) error                           { return nil }
func (c *recordingContext) ShouldBindJSON(any) error                       { return nil }
func (c *recordingContext) ShouldBindQuery(any) error                      { return nil }
func (c *recordingContext) BindJSON(any) error                             { return nil }
func (c *recordingContext) FormFile(string) (*multipart.FileHeader, error) { return nil, nil }
func (c *recordingContext) JSON(status int, body any) {
	c.responseStatus = status
	if payload, ok := body.(map[string]any); ok {
		c.responseBody = payload
		return
	}
	data, _ := json.Marshal(body)
	_ = json.Unmarshal(data, &c.responseBody)
}
func (c *recordingContext) FileAttachment(string, string)            {}
func (c *recordingContext) AbortWithStatusJSON(status int, body any) { c.JSON(status, body) }

func TestPluginRegisterUsesLocaleKeys(t *testing.T) {
	registryRoot := filepath.Join("..", "..", "..")
	if err := corei18n.LoadResourceRoots(registryRoot); err != nil {
		t.Fatalf("LoadResourceRoots returned error: %v", err)
	}

	plugin := New()
	registrar := &recordingRegistrar{}
	if err := plugin.Register(&pluginiface.Context{}, registrar); err != nil {
		t.Fatalf("Register returned error: %v", err)
	}
	if len(registrar.routes) != 1 {
		t.Fatalf("expected 1 route, got %d", len(registrar.routes))
	}
	if len(registrar.menus) != 2 {
		t.Fatalf("expected 2 menus, got %d", len(registrar.menus))
	}
	if len(registrar.permissions) != 1 {
		t.Fatalf("expected 1 permission, got %d", len(registrar.permissions))
	}
	if registrar.menus[0].Name != "示例插件" {
		t.Fatalf("root menu name = %q, want 示例插件", registrar.menus[0].Name)
	}
	if registrar.menus[1].Name != "首页" {
		t.Fatalf("child menu name = %q, want 首页", registrar.menus[1].Name)
	}
	if registrar.permissions[0].Description != "示例插件说明" {
		t.Fatalf("permission description = %q, want 示例插件说明", registrar.permissions[0].Description)
	}

	corei18n.BindRequestLanguage("example-plugin-request", corei18n.LanguageENUS)
	t.Cleanup(func() { corei18n.ClearRequestLanguage("example-plugin-request") })

	ctx := &recordingContext{values: map[string]any{"request_id": "example-plugin-request"}}
	route := registrar.routes[0]
	route.Handler(ctx)
	if ctx.responseStatus != http.StatusOK {
		t.Fatalf("response status = %d, want %d", ctx.responseStatus, http.StatusOK)
	}
	if ctx.responseBody == nil {
		t.Fatal("expected response body")
	}
	if got := ctx.responseBody["message"]; got != "Example plugin response" {
		t.Fatalf("response message = %v, want Example plugin response", got)
	}
}
