package http

import (
	coretransport "goadmin/core/transport"
	"goadmin/modules/dictionary/transport/http/handler"

	"go.uber.org/zap"
)

type Dependencies struct {
	Handler *handler.Handler
	Logger  *zap.Logger
}

func Register(group coretransport.RouteRegistrar, deps Dependencies) {
	if group == nil || deps.Handler == nil {
		return
	}
	root := group.Group("/dictionaries")
	categories := root.Group("/categories")
	categories.GET("", deps.Handler.ListCategories)
	categories.GET("/:id", deps.Handler.GetCategory)
	categories.POST("", deps.Handler.CreateCategory)
	categories.PUT("/:id", deps.Handler.UpdateCategory)
	categories.DELETE("/:id", deps.Handler.DeleteCategory)

	items := root.Group("/items")
	items.GET("", deps.Handler.ListItems)
	items.GET("/:id", deps.Handler.GetItem)
	items.POST("", deps.Handler.CreateItem)
	items.PUT("/:id", deps.Handler.UpdateItem)
	items.DELETE("/:id", deps.Handler.DeleteItem)

	lookup := root.Group("/lookup")
	lookup.GET("/:code", deps.Handler.LookupItems)
	lookup.GET("/:code/:value", deps.Handler.LookupItem)
}
