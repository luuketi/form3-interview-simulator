package main

import (
	tcp_listener "github.com/form3tech-oss/interview-simulator/internal/tcp-listener"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	go tcp_listener.Start()

	<-shutdown
}
