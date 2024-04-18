package http

import (
	"encoding/json"
	"errors"
	"homework/internal/domain"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"nhooyr.io/websocket"
)

type WebSocketHandler struct {
	useCases UseCases

	connections map[*websocket.Conn]struct{}
	m           sync.Mutex
}

func NewWebSocketHandler(useCases UseCases) *WebSocketHandler {
	return &WebSocketHandler{
		useCases:    useCases,
		connections: make(map[*websocket.Conn]struct{}),
		m:           sync.Mutex{},
	}
}

func (h *WebSocketHandler) Handle(ctx *gin.Context, id int64) error {
	_, err := h.useCases.Sensor.GetSensorByID(ctx, id)
	if err != nil {
		return err
	}

	conn, err := websocket.Accept(ctx.Writer, ctx.Request, nil)
	if err != nil {
		return err
	}

	h.m.Lock()
	h.connections[conn] = struct{}{}
	h.m.Unlock()

	go func() {
		c := conn.CloseRead(ctx)
		t := time.NewTicker(time.Second * 2)
		defer t.Stop()

		lastEvent := domain.Event{}
		for {
			select {
			case <-c.Done():
				h.closeConn(conn, websocket.StatusNormalClosure, c.Err().Error())
				return
			case <-t.C:
				event, err := h.useCases.Event.GetLastEventBySensorID(c, id)
				if err != nil {
					h.closeConn(conn, websocket.StatusInternalError, err.Error())
					return
				}
				if lastEvent != *event {
					lastEvent = *event
					js, _ := json.Marshal(event)
					err := conn.Write(c, websocket.MessageText, js)
					if err != nil {
						h.closeConn(conn, websocket.StatusInternalError, err.Error())
						return
					}
				}
			}
		}
	}()

	return nil
}

func (h *WebSocketHandler) closeConn(conn *websocket.Conn, code websocket.StatusCode, reason string) {
	conn.Close(code, reason)

	h.m.Lock()
	delete(h.connections, conn)
	h.m.Unlock()
}

func (h *WebSocketHandler) Shutdown() error {
	h.m.Lock()
	var e []error
	for c := range h.connections {
		e = append(e, c.Close(websocket.StatusNormalClosure, "server shutting down"))
	}
	h.m.Unlock()
	return errors.Join(e...)
}
