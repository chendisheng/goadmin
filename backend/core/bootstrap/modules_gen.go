package bootstrap

import (
	"goadmin/modules/book"
	"goadmin/modules/casbin_model"
	"goadmin/modules/casbin_rule"
)

func generatedModules() []Module {
	return []Module{
		book.NewBootstrap(),
		casbin_model.NewBootstrap(),
		casbin_rule.NewBootstrap(),
	}
}
