package main

import "github.com/gin-gonic/gin"

type Router struct {
	e *gin.Engine
}

func StartRouter(listenAddr string) {
	r := Router{gin.Default()}
	r.e.Run(listenAddr)
}
