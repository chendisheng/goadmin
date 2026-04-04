package http

import coretransport "goadmin/core/transport"

func Register(group coretransport.RouteRegistrar, deps Dependencies) {
	if group == nil {
		return
	}
	h := NewHandler(deps)
	root := group.Group("/codegen")
	dsl := root.Group("/dsl")
	dsl.POST("/preview", h.Preview)
	dsl.POST("/generate", h.Generate)
	dsl.POST("/generate-download", h.GenerateDownload)
	db := root.Group("/db")
	db.POST("/preview", h.PreviewDatabase)
	db.POST("/generate", h.GenerateDatabase)
	db.POST("/generate-download", h.GenerateDatabaseDownload)
	root.GET("/artifacts/:taskID", h.DownloadArtifact)
}
