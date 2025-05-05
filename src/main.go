package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	e := gin.Default()

	e.LoadHTMLGlob("templates/**/*")

	e.GET("/", root)
	e.GET("/home", home)

	e.Run(":42069")
}

func root(c *gin.Context) {
	data := gin.H{"Message": "Hello"}
	c.HTML(http.StatusOK, "index.tmpl", data)
}

func home(c *gin.Context) {
	data := gin.H{"Message": "Hello"}
	c.HTML(http.StatusOK, "home/home.tmpl", data)
}
