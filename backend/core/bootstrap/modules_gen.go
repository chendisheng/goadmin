package bootstrap

import (
	book "goadmin/modules/book"
	codegen_console "goadmin/modules/codegen_console"
	menu "goadmin/modules/menu"
	order "goadmin/modules/order"
	role "goadmin/modules/role"
	user "goadmin/modules/user"
)

func generatedModules() []Module {
	return []Module{
		book.NewBootstrap(),
		codegen_console.NewBootstrap(),
		menu.NewBootstrap(),
		order.NewBootstrap(),
		role.NewBootstrap(),
		user.NewBootstrap(),
	}
}
