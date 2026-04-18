package bootstrap

import (
	"goadmin/modules/casbin_model"
	"goadmin/modules/casbin_rule"
)

func generatedModules() []Module {
	return []Module{
		casbin_model.NewBootstrap(),
		casbin_rule.NewBootstrap(),
	}
}
