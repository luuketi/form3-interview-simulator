package tcp_listener

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/form3tech-oss/interview-simulator/internal/payment"
	response "github.com/form3tech-oss/interview-simulator/internal/response"
	"github.com/rs/zerolog"
	"io"
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

type newScanner interface {
	NewScanner(r io.Reader) Scanner
}

type BufioScanner struct{}

func (s BufioScanner) NewScanner(r io.Reader) Scanner {
	return bufio.NewScanner(r)
}

type Scanner interface {
	Scan() bool
	Text() string
	Err() error
}

type TcpListener struct {
	wg               sync.WaitGroup
	waitPeriod       time.Duration
	mu               sync.Mutex
	connections      map[net.Conn]struct{}
	shutdownListener bool
	listener         net.Listener
	deps             TcpListenerDeps
}

type TcpListenerDeps struct {
	Logger     zerolog.Logger
	Listener   networkListener
	NewScanner newScanner
}

func New(port uint16, waitPeriod time.Duration, deps *TcpListenerDeps) (*TcpListener, error) {
	l, err := deps.Listener.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		deps.Logger.Error().Err(err).Msg("Error listening connection.")
		return nil, err
	}
	return &TcpListener{
			listener:         l,
			deps:             *deps,
			waitPeriod:       waitPeriod,
			connections:      make(map[net.Conn]struct{}),
			shutdownListener: false,
		},
		nil
}

func (l *TcpListener) Start() {
	l.deps.Logger.Info().Msg("Starting service...")
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
				l.deps.Logger.Error().Err(err).Msg("Error accepting connection.")
			}
			continue
		}
		l.deps.Logger.Info().Msg("Accepted new connection.")
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
		l.deps.Logger.Error().Err(err).Msg("Error closing listener.")
		return
	}

	done := make(chan struct{})
	go func() {
		l.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		l.deps.Logger.Info().Msg("All connections completed gracefully.")
	case <-time.After(l.waitPeriod):
		l.deps.Logger.Info().Msg("Grace period finished for active requests. Cancelling pending requests...")
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
			l.deps.Logger.Error().Err(err).Msg("Error sending cancelled response.")
			return
		}
		l.closeConnection(connection)
	}
}

func (l *TcpListener) closeConnection(connection net.Conn) {
	err := connection.Close()
	if err != nil {
		l.deps.Logger.Error().Err(err).Msg("Error closing connection.")
	}
}

func (l *TcpListener) sendResponse(connection net.Conn, resp string) (err error) {
	l.deps.Logger.Debug().Str("response", resp).Msg("Sending response.")
	_, err = fmt.Fprintf(connection, "%s\n", resp)
	if err != nil {
		l.deps.Logger.Error().Err(err).Msg("Error writing response to connection.")
	}
	return
}

func (l *TcpListener) handleConnection(connection net.Conn) {
	defer l.wg.Done()
	defer l.deleteAndCloseConnection(connection)

	scanner := l.deps.NewScanner.NewScanner(connection)
	for scanner.Scan() {
		request := scanner.Text()
		l.deps.Logger.Debug().Str("request", request).Msg("Received request.")
		payment := payment.FromString(request)
		resp := payment.Process()
		if err := l.sendResponse(connection, resp.ToString()); err != nil {
			return
		}
	}
	if err := scanner.Err(); err != nil {
		l.deps.Logger.Error().Err(err).Msg("Error reading from connection.")
	}
}

func (l *TcpListener) deleteAndCloseConnection(connection net.Conn) {
	l.mu.Lock()
	delete(l.connections, connection)
	l.mu.Unlock()
	l.closeConnection(connection)
}
