package http

import (
	"context"
	"errors"
	"fmt"
	"homework/internal/usecase"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Server struct {
	host      string
	port      uint16
	router    *gin.Engine
	wsHandler *WebSocketHandler
}

const (
	DefaultPort = 8080
	DefaultHost = "localhost"
)

type UseCases struct {
	Event  *usecase.Event
	Sensor *usecase.Sensor
	User   *usecase.User
}

func NewServer(useCases UseCases, options ...func(*Server)) *Server {
	r := gin.Default()
	ws := NewWebSocketHandler(useCases)
	setupRouter(r, useCases, ws)

	s := &Server{router: r, host: DefaultHost, port: DefaultPort, wsHandler: ws}
	for _, o := range options {
		o(s)
	}

	return s
}

func WithHost(host string) func(*Server) {
	return func(s *Server) {
		s.host = host
	}
}

func WithPort(port uint16) func(*Server) {
	return func(s *Server) {
		s.port = port
	}
}

func (s *Server) Run(ctx context.Context) error {
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", s.host, s.port),
		Handler: s.router,
	}

	done := make(chan error)
	go func() {
		done <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		c, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		es := []error{server.Shutdown(c), s.wsHandler.Shutdown()}
		return errors.Join(es...)
	case err := <-done:
		return err
	}
}
