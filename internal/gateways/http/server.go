package http

import (
	"context"
	"fmt"
	"homework/internal/usecase"

	"github.com/gin-gonic/gin"
)

type Server struct {
	host   string
	port   uint16
	router *gin.Engine
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
	setupRouter(r, useCases, NewWebSocketHandler(useCases))

	s := &Server{router: r, host: DefaultHost, port: DefaultPort}
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
	return s.router.Run(fmt.Sprintf("%s:%d", s.host, s.port))
}
