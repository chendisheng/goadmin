package loader

import (
	"fmt"
	"sort"
	"strings"

	pluginiface "goadmin/plugin/interface"
	pluginregistry "goadmin/plugin/registry"
)

func Load(ctx *pluginiface.Context, plugins ...pluginiface.Plugin) (*pluginregistry.Registry, error) {
	if ctx == nil {
		ctx = &pluginiface.Context{}
	}
	reg := pluginregistry.New()
	ordered := make([]pluginiface.Plugin, 0, len(plugins))
	for _, plugin := range plugins {
		if plugin == nil {
			continue
		}
		ordered = append(ordered, plugin)
	}
	sort.SliceStable(ordered, func(i, j int) bool {
		return strings.ToLower(ordered[i].Name()) < strings.ToLower(ordered[j].Name())
	})

	for _, plugin := range ordered {
		if err := reg.RegisterPlugin(plugin.Name()); err != nil {
			return nil, err
		}
		if err := plugin.Register(ctx, reg); err != nil {
			return nil, fmt.Errorf("register plugin %q: %w", plugin.Name(), err)
		}
	}

	return reg, nil
}
