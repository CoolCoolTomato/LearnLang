package controllers

import (
	"learnlang-api/websocket"
	"net/http"

	"github.com/gin-gonic/gin"
	ws "github.com/gorilla/websocket"
)

var upgrader = ws.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type WebSocketController struct {
	hub *websocket.Hub
}

func NewWebSocketController(hub *websocket.Hub) *WebSocketController {
	return &WebSocketController{hub: hub}
}

func (wsc *WebSocketController) HandleWebSocket(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := &websocket.Client{
		UserID: userID.(int64),
		Conn:   conn,
		Send:   make(chan []byte, 256),
	}

	wsc.hub.Register(client)

	go client.WritePump()
	go client.ReadPump(wsc.hub)
}
