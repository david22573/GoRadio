package app

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(a *App) {
	a.Router.GET("/", func(ctx *gin.Context) { ctx.HTML(200, "index.tmpl", nil) })
	a.Router.GET("/ping", func(ctx *gin.Context) { ctx.JSON(200, gin.H{"message": "pong"}) })
}
