package bootstrap

import (
	"goadmin/modules/codegen_console"
	"goadmin/modules/dictionary"
	"goadmin/modules/menu"
	"goadmin/modules/role"
	"goadmin/modules/user"
)

func builtinModules() []Module {
	return []Module{
		codegen_console.NewBootstrap(),
		dictionary.NewBootstrap(),
		menu.NewBootstrap(),
		role.NewBootstrap(),
		user.NewBootstrap(),
	}
}

func BuiltinModules() []Module {
	return builtinModules()
}
