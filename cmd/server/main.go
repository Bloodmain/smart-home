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
)

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
	if !present {
		port = httpGateway.DefaultPort
	}
	if err != nil || port < 0 || port > 9999 {
		log.Fatalf("invalid port number: %s\n", portRaw)
	}

	r := httpGateway.NewServer(useCases, httpGateway.WithHost(host), httpGateway.WithPort(uint16(port)))
	if err := r.Run(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Printf("error during server shutdown: %v", err)
	}
}
