package tcp_listener

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type TcpListenerTestSuite struct {
	suite.Suite
}

func TestTcpListenerSuite(t *testing.T) {
	tcpListenerTestSuite := &TcpListenerTestSuite{}

	go Start()

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
