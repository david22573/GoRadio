package app

import "github.com/gin-gonic/gin"

type Router struct {
	*gin.Engine
	listenAddr string
}

func NewRouter() *Router {
	return &Router{Engine: gin.Default()}
}
