package tcp_listener

import (
	"bufio"
	"fmt"
	"github.com/form3tech-oss/interview-simulator/internal/payment"
	"github.com/rs/zerolog"
	"net"
	"time"
)

type networkListener interface {
	Listen(network string, address string) (net.Listener, error)
}

type NetListener struct{}

func (s NetListener) Listen(network string, address string) (net.Listener, error) {
	return net.Listen(network, address)
}

type TcpListener struct {
	logger     zerolog.Logger
	listener   net.Listener
	waitPeriod time.Duration
}

func New(logger zerolog.Logger, listener networkListener, port uint16, waitPeriod time.Duration) (*TcpListener, error) {
	l, err := listener.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		logger.Error().Err(err).Msg("Error listening connection.")
		return nil, err
	}
	return &TcpListener{
			logger:     logger,
			listener:   l,
			waitPeriod: waitPeriod,
		},
		nil
}

func (l *TcpListener) Start() error {
	defer l.listener.Close()
	l.logger.Info().Msg("Starting service...")

	for {
		conn, err := l.listener.Accept()
		if err != nil {
			l.logger.Error().Err(err).Msg("Error accepting connection.")
			continue
		}

		go l.handleConnection(conn)
	}
}

func (l *TcpListener) handleConnection(conn net.Conn) {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		request := scanner.Text()
		response := l.handleRequest(request)
		fmt.Fprintf(conn, "%s\n", response)
	}

	if err := scanner.Err(); err != nil {
		l.logger.Error().Err(err).Msg("Error reading from connection.")
	}
}

func (l *TcpListener) handleRequest(request string) string {
	payment := payment.FromString(request)
	response := payment.Process()
	return response.ToString()
}
