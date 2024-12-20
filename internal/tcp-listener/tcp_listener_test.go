package tcp_listener

import (
	"bufio"
	"fmt"
	"github.com/rs/zerolog"
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

type TcpListenerTestSuite struct {
	suite.Suite
}

func TestTcpListenerSuite(t *testing.T) {
	tcpListenerTestSuite := &TcpListenerTestSuite{}

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	listener, err := New(logger, PORT, WAIT_PERIOD)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	go listener.Start()

	// wait of the server to be ready
	time.Sleep(time.Second)

	suite.Run(t, tcpListenerTestSuite)
}

func (suite *TcpListenerTestSuite) TestSchemeSimulator() {
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
			conn, err := net.Dial("tcp", ":8080")
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

func (suite *TcpListenerTestSuite) Test_TwoRequestsInOneConnection() {
	msg1 := "PAYMENT|10"
	msg2 := "PAYMENT|50"
	expectedResponse := "RESPONSE|ACCEPTED|Transaction processed"

	conn, err := net.Dial("tcp", ":8080")
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

	response2, err := bufio.NewReader(conn).ReadString('\n')
	suite.NoError(err, "Failed to read response")
	response2 = strings.TrimSpace(response2)

	duration := time.Since(start)

	suite.Equal(expectedResponse, response1, "Unexpected response")
	suite.Equal(expectedResponse, response2, "Unexpected response")

	suite.LessOrEqual(duration, 50*time.Millisecond, "Response time was longer than expected")
}

func (suite *TcpListenerTestSuite) Test_TwoRequestsInTwoConnections() {
	msg1 := "PAYMENT|10"
	msg2 := "PAYMENT|50"
	expectedResponse := "RESPONSE|ACCEPTED|Transaction processed"

	conn1, err := net.Dial("tcp", ":8080")
	suite.NoError(err, "Failed to connect to server")
	defer conn1.Close()

	conn2, err := net.Dial("tcp", ":8080")
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
