package tcp_listener

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/form3tech-oss/interview-simulator/internal/payment"
	response "github.com/form3tech-oss/interview-simulator/internal/response"
	"github.com/rs/zerolog"
	"net"
	"sync"
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
	logger           zerolog.Logger
	wg               sync.WaitGroup
	listener         net.Listener
	waitPeriod       time.Duration
	mu               sync.Mutex
	connections      map[net.Conn]struct{}
	shutdownListener bool
}

func New(logger zerolog.Logger, listener networkListener, port uint16, waitPeriod time.Duration) (*TcpListener, error) {
	l, err := listener.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		logger.Error().Err(err).Msg("Error listening connection.")
		return nil, err
	}
	return &TcpListener{
			logger:           logger,
			listener:         l,
			waitPeriod:       waitPeriod,
			connections:      make(map[net.Conn]struct{}),
			shutdownListener: false,
		},
		nil
}

func (l *TcpListener) Start() {
	l.logger.Info().Msg("Starting service...")
	for {
		connection, err := l.listener.Accept()
		if err != nil {
			l.mu.Lock()
			if l.shutdownListener {
				l.mu.Unlock()
				break
			}
			l.mu.Unlock()
			if !errors.Is(err, net.ErrClosed) {
				l.logger.Error().Err(err).Msg("Error accepting connection.")
			}
			continue
		}
		l.logger.Info().Msg("Accepted new connection.")
		l.storeConnection(connection)
		go l.handleConnection(connection)
	}
}

func (l *TcpListener) Stop() {
	l.mu.Lock()
	l.shutdownListener = true
	err := l.listener.Close()
	l.mu.Unlock()
	if err != nil {
		l.logger.Error().Err(err).Msg("Error closing listener.")
		return
	}

	done := make(chan struct{})
	go func() {
		l.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		l.logger.Info().Msg("All connections completed gracefully.")
	case <-time.After(l.waitPeriod):
		l.logger.Info().Msg("Grace period finished for active requests. Cancelling pending requests...")
		l.closeConnections()
	}
}

func (l *TcpListener) storeConnection(conn net.Conn) {
	l.wg.Add(1)
	l.mu.Lock()
	l.connections[conn] = struct{}{}
	l.mu.Unlock()
}

func (l *TcpListener) closeConnections() {
	defer l.mu.Unlock()
	l.mu.Lock()
	for connection := range l.connections {
		rejected := response.NewRejected("Cancelled")
		if err := l.sendResponse(connection, rejected.ToString()); err != nil {
			return
		}
		connection.Close()
	}
}

func (l *TcpListener) sendResponse(connection net.Conn, resp string) (err error) {
	l.logger.Debug().Str("response", resp).Msg("Sending response.")
	_, err = fmt.Fprintf(connection, "%s\n", resp)
	if err != nil {
		l.logger.Error().Err(err).Msg("Error writing response to connection.")
	}
	return
}

func (l *TcpListener) handleConnection(connection net.Conn) {
	defer l.wg.Done()
	defer l.deleteAndCloseConnection(connection)

	scanner := bufio.NewScanner(connection)
	for scanner.Scan() {
		request := scanner.Text()
		l.logger.Debug().Str("request", request).Msg("Received request.")
		payment := payment.FromString(request)
		resp := payment.Process()
		if err := l.sendResponse(connection, resp.ToString()); err != nil {
			return
		}
	}
	if err := scanner.Err(); err != nil {
		l.logger.Error().Err(err).Msg("Error reading from connection.")
	}
}

func (l *TcpListener) deleteAndCloseConnection(connection net.Conn) {
	l.mu.Lock()
	delete(l.connections, connection)
	l.mu.Unlock()
	connection.Close()
}
