package tcp_listener

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

const (
	PORT        = 8080
	WAIT_PERIOD = 5 * time.Second
)

func TestMain(m *testing.M) {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	listener, err := New(logger, NetListener{}, PORT, WAIT_PERIOD)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	go listener.Start()

	// wait of the server to be ready
	time.Sleep(time.Second)

	code := m.Run()

	os.Exit(code)
}

func TestSchemeSimulator(t *testing.T) {
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
		t.Run(tt.name, func(t *testing.T) {
			conn, err := net.Dial("tcp", ":8080")
			require.NoError(t, err, "Failed to connect to server")
			defer conn.Close()

			_, err = fmt.Fprintf(conn, tt.input+"\n")
			require.NoError(t, err, "Failed to send request")

			start := time.Now()

			response, err := bufio.NewReader(conn).ReadString('\n')
			require.NoError(t, err, "Failed to read response")
			duration := time.Since(start)

			response = strings.TrimSpace(response)

			require.Equal(t, tt.expectedOutput, response, "Unexpected response")

			if tt.minDuration > 0 {
				require.GreaterOrEqual(t, duration, tt.minDuration, "Response time was shorter than expected")
			}

			if tt.maxDuration > 0 {
				require.LessOrEqual(t, duration, tt.maxDuration, "Response time was longer than expected")
			}
		})
	}
}

func Test_TwoRequestsInOneConnection(t *testing.T) {
	msg1 := "PAYMENT|10"
	msg2 := "PAYMENT|50"
	expectedResponse := "RESPONSE|ACCEPTED|Transaction processed"

	conn, err := net.Dial("tcp", ":8080")
	require.NoError(t, err, "Failed to connect to server")
	defer conn.Close()

	_, err = fmt.Fprintf(conn, msg1+"\n")
	require.NoError(t, err, "Failed to send request 1")
	_, err = fmt.Fprintf(conn, msg2+"\n")
	require.NoError(t, err, "Failed to send request 2")

	start := time.Now()

	response1, err := bufio.NewReader(conn).ReadString('\n')
	require.NoError(t, err, "Failed to read response")
	response1 = strings.TrimSpace(response1)

	response2, err := bufio.NewReader(conn).ReadString('\n')
	require.NoError(t, err, "Failed to read response")
	response2 = strings.TrimSpace(response2)

	duration := time.Since(start)

	require.Equal(t, expectedResponse, response1, "Unexpected response")
	require.Equal(t, expectedResponse, response2, "Unexpected response")

	require.LessOrEqual(t, duration, 50*time.Millisecond, "Response time was longer than expected")
}

func Test_TwoRequestsInTwoConnections(t *testing.T) {
	msg1 := "PAYMENT|10"
	msg2 := "PAYMENT|50"
	expectedResponse := "RESPONSE|ACCEPTED|Transaction processed"

	conn1, err := net.Dial("tcp", ":8080")
	require.NoError(t, err, "Failed to connect to server")
	defer conn1.Close()

	conn2, err := net.Dial("tcp", ":8080")
	require.NoError(t, err, "Failed to connect to server")
	defer conn2.Close()

	_, err = fmt.Fprintf(conn1, msg1+"\n")
	require.NoError(t, err, "Failed to send request 1")
	_, err = fmt.Fprintf(conn2, msg2+"\n")
	require.NoError(t, err, "Failed to send request 2")

	start := time.Now()

	response1, err := bufio.NewReader(conn1).ReadString('\n')
	require.NoError(t, err, "Failed to read response")
	response1 = strings.TrimSpace(response1)

	response2, err := bufio.NewReader(conn2).ReadString('\n')
	require.NoError(t, err, "Failed to read response")
	response2 = strings.TrimSpace(response2)

	duration := time.Since(start)

	require.Equal(t, expectedResponse, response1, "Unexpected response")
	require.Equal(t, expectedResponse, response2, "Unexpected response")

	require.LessOrEqual(t, duration, 50*time.Millisecond, "Response time was longer than expected")
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
}

func (m *ListenerMock) Accept() (net.Conn, error) {
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
