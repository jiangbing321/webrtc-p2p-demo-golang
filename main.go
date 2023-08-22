package main

import (
	"fmt"

	"webrtc.webrtc-p2p-demo-go/websocket"

	"github.com/gin-gonic/gin"
)

func main() {
	fmt.Println("hello world")
	r := gin.Default()
	r.Static("/static", "./static")
	r.GET("/websocket", websocket.Server)
	r.Run(":5001")
}
