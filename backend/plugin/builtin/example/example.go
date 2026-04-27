package exampleplugin

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	corei18n "goadmin/core/i18n"
	coretransport "goadmin/core/transport"
	pluginiface "goadmin/plugin/interface"

	"go.uber.org/zap"
)

type Plugin struct{}

func New() *Plugin {
	return &Plugin{}
}

func (p *Plugin) Name() string {
	return "example"
}

func (p *Plugin) Register(ctx *pluginiface.Context, registrar pluginiface.Registrar) error {
	if registrar == nil {
		return fmt.Errorf("plugin registrar is required")
	}
	if ctx != nil && ctx.Logger != nil {
		ctx.Logger.Info("register example plugin", zap.String("title", resolveLocaleText("plugin.example.title", "Plugin Example")))
	}
	pluginTitle := resolveLocaleText("plugin.example.title", "Plugin Example")
	pluginHomeTitle := resolveLocaleText("plugin.example.home", "Home")
	pluginDescription := resolveLocaleText("plugin.example.description", "Example plugin description")
	pluginResponse := resolveLocaleText("plugin.example.ping", "pong from example plugin")

	if err := registrar.AddRoute(pluginiface.Route{
		Name:   "examplePing",
		Method: http.MethodGet,
		Path:   "/plugins/example/ping",
		Access: pluginiface.AccessPublic,
		Handler: func(c coretransport.Context) {
			requestID := requestID(c)
			c.JSON(http.StatusOK, map[string]any{
				"message": translateWithRequest(requestID, "plugin.example.ping", pluginResponse),
				"plugin":  "example",
			})
		},
	}); err != nil {
		return err
	}

	if err := registrar.AddMenu(pluginiface.Menu{
		ID:           "plugin-example-root",
		Name:         pluginTitle,
		TitleKey:     "route.plugin_example",
		TitleDefault: "Plugin example",
		Path:         "/plugin/example",
		Component:    "Layout",
		Icon:         "plug",
		Sort:         100,
		Permission:   "plugin:example:view",
		Type:         pluginiface.MenuTypeDirectory,
		Visible:      true,
		Enabled:      true,
		Redirect:     "/plugin/example/home",
	}); err != nil {
		return err
	}

	if err := registrar.AddMenu(pluginiface.Menu{
		ID:           "plugin-example-home",
		ParentID:     "plugin-example-root",
		Name:         pluginHomeTitle,
		TitleKey:     "route.plugin_example_home",
		TitleDefault: "Home",
		Path:         "/plugin/example/home",
		Component:    "view/plugin/example/index",
		Icon:         "sparkles",
		Sort:         1,
		Permission:   "plugin:example:view",
		Type:         pluginiface.MenuTypeMenu,
		Visible:      true,
		Enabled:      true,
	}); err != nil {
		return err
	}

	return registrar.AddPermission(pluginiface.Permission{
		Object:      "plugin:example",
		Action:      "view",
		Description: pluginDescription,
	})
}

func requestID(c coretransport.Context) string {
	if c == nil {
		return ""
	}
	if value, exists := c.Get("request_id"); exists {
		if id, ok := value.(string); ok && strings.TrimSpace(id) != "" {
			return id
		}
	}
	return ""
}

func resolveLocaleText(key, fallback string) string {
	if translated := corei18n.DefaultRegistry().MustTranslate(context.Background(), key); translated != key {
		return translated
	}
	if strings.TrimSpace(fallback) != "" {
		return fallback
	}
	return key
}

func translateWithRequest(requestID, key, fallback string) string {
	if translated := corei18n.TranslateRequest(requestID, key); translated != key {
		return translated
	}
	if strings.TrimSpace(fallback) != "" {
		return fallback
	}
	return key
}
