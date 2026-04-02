package http

import coretransport "goadmin/core/transport"

func Register(group coretransport.RouteRegistrar, deps Dependencies) {
	if group == nil {
		return
	}
	h := NewHandler(deps)
	routes := group.Group("/codegen").Group("/dsl")
	routes.POST("/preview", h.Preview)
	routes.POST("/generate", h.Generate)
}
