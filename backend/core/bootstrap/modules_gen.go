package bootstrap

import (
	"goadmin/modules/book"
	"goadmin/modules/order"
)

func generatedModules() []Module {
	return []Module{
		book.NewBootstrap(),
		order.NewBootstrap(),
	}
}
