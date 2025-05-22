package app

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(a *App) {
	a.Router.GET("/radio", func(ctx *gin.Context) { ctx.HTML(200, "radio/index.tmpl", nil) })
	a.Router.GET("/", func(ctx *gin.Context) { ctx.HTML(200, "index.tmpl", nil) })
}
