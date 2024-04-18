package http

import (
	"github.com/gin-gonic/gin"
)

type WebSocketHandler struct {
	useCases UseCases
}

func NewWebSocketHandler(useCases UseCases) *WebSocketHandler {
	return &WebSocketHandler{
		useCases: useCases,
	}
}

func (h *WebSocketHandler) Handle(c *gin.Context, id int64) error {
	// TODO implement me
	return nil
}

func (h *WebSocketHandler) Shutdown() error {
	// TODO implement me
	return nil
}
