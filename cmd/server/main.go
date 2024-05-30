package main

import (
	"context"
	"errors"
	"homework/internal/usecase"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"

	httpGateway "homework/internal/gateways/http"
	eventRepository "homework/internal/repository/event/inmemory"
	sensorRepository "homework/internal/repository/sensor/inmemory"
	userRepository "homework/internal/repository/user/inmemory"

	_ "github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	MetricsAddr = "/metrics"
	MetricsPort = 8000
)

func runMetrics() {
	metricServer := http.NewServeMux()
	metricServer.Handle(MetricsAddr, promhttp.Handler())

	log.Printf("Listening metrics on :%d", MetricsPort)
	err := http.ListenAndServe(":"+strconv.Itoa(MetricsPort), metricServer)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	er := eventRepository.NewEventRepository()
	sr := sensorRepository.NewSensorRepository()
	ur := userRepository.NewUserRepository()
	sor := userRepository.NewSensorOwnerRepository()

	useCases := httpGateway.UseCases{
		Event:  usecase.NewEvent(er, sr),
		Sensor: usecase.NewSensor(sr),
		User:   usecase.NewUser(ur, sor, sr),
	}

	host, present := os.LookupEnv("HTTP_HOST")
	if !present {
		host = httpGateway.DefaultHost
	}
	portRaw, present := os.LookupEnv("HTTP_PORT")
	port, err := strconv.Atoi(portRaw)
	if !present || err != nil || port < 0 || port > 9999 {
		log.Printf("Valid port number hasn't been provided, using default port")
		port = httpGateway.DefaultPort
	}

	if port == MetricsPort {
		log.Fatalf("Port number clashes with metrics port: %s\n", portRaw)
		return
	}

	go runMetrics()

	r := httpGateway.NewServer(useCases, httpGateway.WithHost(host), httpGateway.WithPort(uint16(port)))
	if err := r.Run(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Printf("error during server shutdown: %v", err)
	}
}
