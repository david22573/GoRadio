package app

import "github.com/gin-gonic/gin"

func RegisterRoutes(a *App) {
	a.Router.GET("/ping", func(ctx *gin.Context) { ctx.JSON(200, gin.H{"message": "pong"}) })
}
