package bootstrap

import (
	"goadmin/modules/codegen_console"
	"goadmin/modules/menu"
	"goadmin/modules/role"
	"goadmin/modules/user"
)

func builtinModules() []Module {
	return []Module{
		codegen_console.NewBootstrap(),
		menu.NewBootstrap(),
		role.NewBootstrap(),
		user.NewBootstrap(),
	}
}
