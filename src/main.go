package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	rc := NewRadioClient("https://kxlu.streamguys1.com/kxlu-lo")
	rc.Record(10, "test.mp3")

	e := gin.Default()

	e.Run(":42069")
}
