package exampleplugin

import (
	"fmt"
	"net/http"

	coretransport "goadmin/core/transport"
	pluginiface "goadmin/plugin/interface"
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
		ctx.Logger.Info("register example plugin")
	}

	if err := registrar.AddRoute(pluginiface.Route{
		Name:   "examplePing",
		Method: http.MethodGet,
		Path:   "/plugins/example/ping",
		Access: pluginiface.AccessPublic,
		Handler: func(c coretransport.Context) {
			c.JSON(http.StatusOK, map[string]any{
				"message": "pong from example plugin",
				"plugin":  "example",
			})
		},
	}); err != nil {
		return err
	}

	if err := registrar.AddMenu(pluginiface.Menu{
		ID:         "plugin-example-root",
		Name:       "Plugin Example",
		Path:       "/plugin/example",
		Component:  "Layout",
		Icon:       "plug",
		Sort:       100,
		Permission: "plugin:example:view",
		Type:       pluginiface.MenuTypeDirectory,
		Visible:    true,
		Enabled:    true,
		Redirect:   "/plugin/example/home",
	}); err != nil {
		return err
	}

	if err := registrar.AddMenu(pluginiface.Menu{
		ID:         "plugin-example-home",
		ParentID:   "plugin-example-root",
		Name:       "Home",
		Path:       "/plugin/example/home",
		Component:  "view/plugin/example/index",
		Icon:       "sparkles",
		Sort:       1,
		Permission: "plugin:example:view",
		Type:       pluginiface.MenuTypeMenu,
		Visible:    true,
		Enabled:    true,
	}); err != nil {
		return err
	}

	return registrar.AddPermission(pluginiface.Permission{
		Object:      "plugin:example",
		Action:      "view",
		Description: "View example plugin",
	})
}
