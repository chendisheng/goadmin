package http

import (
	coretransport "goadmin/core/transport"

	"go.uber.org/zap"
)

type Dependencies struct {
	Logger *zap.Logger
}

func Register(group coretransport.RouteRegistrar, deps Dependencies) {
	if group == nil {
		return
	}
	root := group.Group("/dictionaries")
	root.Group("/categories")
	root.Group("/items")
	root.Group("/lookup")
}
