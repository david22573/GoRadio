package api

import "github.com/gin-gonic/gin"

func RegisterHandlers(r *Router) {
	r.GET("/", func(ctx *gin.Context) { ctx.File("templates/index.html") })
	api := r.Group("/api")
	{
		api.GET("/health", func(ctx *gin.Context) { ctx.String(200, "OK") })
	}
}
