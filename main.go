package main

import (
	tcp_listener "github.com/form3tech-oss/interview-simulator/internal/tcp-listener"
	"github.com/rs/zerolog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	PORT        = 8080
	WAIT_PERIOD = 5 * time.Second
)

func main() {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	listener, err := tcp_listener.New(PORT, WAIT_PERIOD, &tcp_listener.TcpListenerDeps{
		Logger:     logger,
		Listener:   tcp_listener.NetListener{},
		NewScanner: tcp_listener.BufioScanner{},
	})
	if err != nil {
		logger.Error().Err(err).Msg("Error creating listener.")
		os.Exit(1)
	}

	go listener.Start()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)
	<-shutdown

	logger.Info().Msg("Shutting down service...")
	listener.Stop()
	logger.Info().Msg("Service stopped.")
}
