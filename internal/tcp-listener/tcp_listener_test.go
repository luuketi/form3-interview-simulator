package tcp_listener

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"math/rand"
	"net"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

const (
	WAIT_PERIOD = 5 * time.Second
)

type NetListenTestSuite struct {
	suite.Suite
	listener *TcpListener
	port     uint16
}

func TestNetListenSuite(t *testing.T) {
	mainTestSuite := &NetListenTestSuite{}
	suite.Run(t, mainTestSuite)
}

func (suite *NetListenTestSuite) SetupTest() {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	port := rndPort()
	listener, err := New(logger, NetListener{}, port, WAIT_PERIOD)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	suite.listener = listener
	suite.port = port

	go suite.listener.Start()

	// wait of the server to be ready
	time.Sleep(time.Second)
}

func (suite *NetListenTestSuite) TearDownTest() {
	suite.listener.Stop()
}

func (suite *NetListenTestSuite) TestSchemeSimulator() {
	tests := []struct {
		name           string
		input          string
		expectedOutput string
		minDuration    time.Duration
		maxDuration    time.Duration
	}{
		{
			name:           "Valid Request",
			input:          "PAYMENT|10",
			expectedOutput: "RESPONSE|ACCEPTED|Transaction processed",
			maxDuration:    50 * time.Millisecond,
		},
		{
			name:           "Valid Request with Delay",
			input:          "PAYMENT|101",
			expectedOutput: "RESPONSE|ACCEPTED|Transaction processed",
			minDuration:    101 * time.Millisecond,
			maxDuration:    151 * time.Millisecond,
		},
		{
			name:           "Invalid Amount with negative number",
			input:          "PAYMENT|-101",
			expectedOutput: "RESPONSE|REJECTED|Invalid amount",
			maxDuration:    10 * time.Millisecond,
		},
		{
			name:           "Invalid Amount with decimal numbers",
			input:          "PAYMENT|101.123",
			expectedOutput: "RESPONSE|REJECTED|Invalid amount",
			maxDuration:    10 * time.Millisecond,
		},
		{
			name:           "Empty Amount",
			input:          "PAYMENT|",
			expectedOutput: "RESPONSE|REJECTED|Invalid amount",
			maxDuration:    10 * time.Millisecond,
		},
		{
			name:           "Invalid Request Format",
			input:          "INVALID|100",
			expectedOutput: "RESPONSE|REJECTED|Invalid request",
			maxDuration:    10 * time.Millisecond,
		},
		{
			name:           "Invalid Request Format with extra field",
			input:          "PAYMENT|10|HELLO",
			expectedOutput: "RESPONSE|REJECTED|Invalid request",
			maxDuration:    10 * time.Millisecond,
		},
		{
			name:           "Large Amount",
			input:          "PAYMENT|20000",
			expectedOutput: "RESPONSE|ACCEPTED|Transaction processed",
			minDuration:    10 * time.Second,
			maxDuration:    10*time.Second + 50*time.Millisecond,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			conn, err := net.Dial("tcp", fmt.Sprintf(":%d", suite.port))
			suite.NoError(err, "Failed to connect to server")
			defer conn.Close()

			_, err = fmt.Fprintf(conn, tt.input+"\n")
			suite.NoError(err, "Failed to send request")

			start := time.Now()

			response, err := bufio.NewReader(conn).ReadString('\n')
			suite.NoError(err, "Failed to read response")
			duration := time.Since(start)

			response = strings.TrimSpace(response)

			suite.Equal(tt.expectedOutput, response, "Unexpected response")

			if tt.minDuration > 0 {
				suite.GreaterOrEqual(duration, tt.minDuration, "Response time was shorter than expected")
			}

			if tt.maxDuration > 0 {
				suite.LessOrEqual(duration, tt.maxDuration, "Response time was longer than expected")
			}
		})
	}
}

func (suite *NetListenTestSuite) Test_TwoRequestsInOneConnection() {
	msg1 := "PAYMENT|10000"
	msg2 := "PAYMENT|-50"
	expectedResponse1 := "RESPONSE|ACCEPTED|Transaction processed"
	expectedResponse2 := "RESPONSE|REJECTED|Invalid amount"

	conn, err := net.Dial("tcp", fmt.Sprintf(":%d", suite.port))
	suite.NoError(err, "Failed to connect to server")
	defer conn.Close()

	_, err = fmt.Fprintf(conn, msg1+"\n")
	suite.NoError(err, "Failed to send request 1")
	_, err = fmt.Fprintf(conn, msg2+"\n")
	suite.NoError(err, "Failed to send request 2")

	start := time.Now()

	response1, err := bufio.NewReader(conn).ReadString('\n')
	suite.NoError(err, "Failed to read response")
	response1 = strings.TrimSpace(response1)

	firstResponseTime := time.Now()

	response2, err := bufio.NewReader(conn).ReadString('\n')
	suite.NoError(err, "Failed to read response")
	response2 = strings.TrimSpace(response2)

	secondResponseTime := time.Now()

	suite.Equal(expectedResponse1, response1, "Unexpected response")
	suite.Equal(expectedResponse2, response2, "Unexpected response")

	suite.LessOrEqual(firstResponseTime.Sub(start), 10*time.Second+50*time.Millisecond, "Response time was longer than expected")
	suite.LessOrEqual(secondResponseTime.Sub(firstResponseTime), 50*time.Millisecond, "Response time was longer than expected")
}

func (suite *NetListenTestSuite) Test_TwoRequestsInTwoConnections() {
	msg1 := "PAYMENT|10"
	msg2 := "PAYMENT|50"
	expectedResponse := "RESPONSE|ACCEPTED|Transaction processed"

	conn1, err := net.Dial("tcp", fmt.Sprintf(":%d", suite.port))
	suite.NoError(err, "Failed to connect to server")
	defer conn1.Close()

	conn2, err := net.Dial("tcp", fmt.Sprintf(":%d", suite.port))
	suite.NoError(err, "Failed to connect to server")
	defer conn2.Close()

	_, err = fmt.Fprintf(conn1, msg1+"\n")
	suite.NoError(err, "Failed to send request 1")
	_, err = fmt.Fprintf(conn2, msg2+"\n")
	suite.NoError(err, "Failed to send request 2")

	start := time.Now()

	response1, err := bufio.NewReader(conn1).ReadString('\n')
	suite.NoError(err, "Failed to read response")
	response1 = strings.TrimSpace(response1)

	response2, err := bufio.NewReader(conn2).ReadString('\n')
	suite.NoError(err, "Failed to read response")
	response2 = strings.TrimSpace(response2)

	duration := time.Since(start)

	suite.Equal(expectedResponse, response1, "Unexpected response")
	suite.Equal(expectedResponse, response2, "Unexpected response")

	suite.LessOrEqual(duration, 50*time.Millisecond, "Response time was longer than expected")
}

func (suite *NetListenTestSuite) Test_CancelRequestDueToGracePeriodExpiration() {
	msg1 := "PAYMENT|50000"
	expectedResponse := "RESPONSE|REJECTED|Cancelled"

	conn, err := net.Dial("tcp", fmt.Sprintf(":%d", suite.port))
	suite.NoError(err, "Failed to connect to server")
	defer conn.Close()

	_, err = fmt.Fprintf(conn, msg1+"\n")
	suite.NoError(err, "Failed to send request 1")

	go suite.listener.Stop()

	start := time.Now()

	response, err := bufio.NewReader(conn).ReadString('\n')
	suite.NoError(err, "Failed to read response")
	response = strings.TrimSpace(response)

	duration := time.Since(start)

	suite.Equal(expectedResponse, response, "Unexpected response")

	suite.LessOrEqual(duration, WAIT_PERIOD+50*time.Millisecond, "Response time was longer than expected")
}

func (suite *NetListenTestSuite) Test_StoppingServiceStopsNewConnections() {
	msg1 := "PAYMENT|50000"

	conn, err := net.Dial("tcp", fmt.Sprintf(":%d", suite.port))
	suite.NoError(err, "Failed to connect to server")
	defer conn.Close()

	_, err = fmt.Fprintf(conn, msg1+"\n")
	suite.NoError(err, "Failed to send request 1")

	go suite.listener.Stop()

	time.Sleep(1 * time.Second)

	conn2, err := net.Dial("tcp", fmt.Sprintf(":%d", suite.port))
	suite.Error(err, "Failed to connect to server")
	suite.Nil(conn2, "Connection should be nil")
}

func (suite *NetListenTestSuite) Test_StoppingServiceKeepsReceivingRequests() {
	msg1 := "PAYMENT|50000"
	msg2 := "PAYMENT|50"
	expectedResponse := "RESPONSE|ACCEPTED|Transaction processed"

	conn, err := net.Dial("tcp", fmt.Sprintf(":%d", suite.port))
	suite.NoError(err, "Failed to connect to server")
	defer conn.Close()

	conn2, err := net.Dial("tcp", fmt.Sprintf(":%d", suite.port))
	suite.NoError(err, "Failed to connect to server")
	defer conn2.Close()

	_, err = fmt.Fprintf(conn, msg1+"\n")
	suite.NoError(err, "Failed to send request 1")

	go suite.listener.Stop()

	time.Sleep(1 * time.Second)

	_, err = fmt.Fprintf(conn2, msg2+"\n")
	suite.NoError(err, "Failed to send request 1")

	response, err := bufio.NewReader(conn2).ReadString('\n')
	suite.NoError(err, "Failed to read response")
	response = strings.TrimSpace(response)
	suite.Equal(expectedResponse, response, "Unexpected response")

	_, err = fmt.Fprintf(conn2, msg2+"\n")
	suite.NoError(err, "Failed to send request 1")

	response2, err := bufio.NewReader(conn2).ReadString('\n')
	suite.NoError(err, "Failed to read response")
	response2 = strings.TrimSpace(response2)
	suite.Equal(expectedResponse, response2, "Unexpected response")
}

type TcpListenerTestSuite struct {
	suite.Suite
	listener *TcpListener
}

func TestTcpListenerSuite(t *testing.T) {
	mainTestSuite := &TcpListenerTestSuite{}
	suite.Run(t, mainTestSuite)
}

func (suite *TcpListenerTestSuite) Test_FailingListener() {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	l := ListenerMock{}
	l.On("Accept").Return(new(net.TCPConn), nil)
	netListener := NetListenerMock(&l)
	expectedErr := errors.New("test error")
	netListener.SetError(expectedErr)

	listener, err := New(logger, &netListener, 8080, WAIT_PERIOD)

	suite.Nil(listener)
	suite.Equal(expectedErr, err)
}

func (suite *TcpListenerTestSuite) Test_FailingAcceptShouldRetry() {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	l := NewListenerMock()
	l.On("Accept").Return(new(net.TCPConn), errors.New("test error")).Times(2)
	netListener := NetListenerMock(l)

	listener, err := New(logger, &netListener, 8080, WAIT_PERIOD)
	go listener.Start()
	time.Sleep(1 * time.Second)

	suite.NotNil(listener)
	suite.Nil(err)
	l.AssertExpectations(suite.T())
}

type MockNetListener struct {
	listener net.Listener
	err      error
}

func NetListenerMock(listener net.Listener) MockNetListener {
	return MockNetListener{listener: listener, err: nil}
}

func (b *MockNetListener) SetError(err error) {
	b.err = err
}

func (b *MockNetListener) Listen(network string, address string) (net.Listener, error) {
	if b.err != nil {
		return nil, b.err
	}
	return b.listener, nil
}

type ListenerMock struct {
	mock.Mock
	wg               sync.WaitGroup
	mu               sync.Mutex
	connectionsCount int
}

func NewListenerMock() *ListenerMock {
	l := &ListenerMock{wg: sync.WaitGroup{}}
	l.wg.Add(1)
	return l
}

func (m *ListenerMock) Accept() (net.Conn, error) {
	m.mu.Lock()
	if m.connectionsCount < 2 {
		m.connectionsCount++
		m.mu.Unlock()
	} else {
		m.wg.Wait()
	}
	args := m.Called()
	return args.Get(0).(net.Conn), args.Error(1)
}

func (m *ListenerMock) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *ListenerMock) Addr() net.Addr {
	args := m.Called()
	return args.Get(0).(net.Addr)
}

func rndPort() uint16 {
	return uint16(rand.Intn(65536-10000) + 10000)
}
