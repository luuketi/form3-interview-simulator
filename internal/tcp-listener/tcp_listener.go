package tcp_listener

import (
	"bufio"
	"fmt"
	"github.com/form3tech-oss/interview-simulator/internal/payment"
	"net"
)

const (
	PORT = 8080
)

func Start() error {
	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", PORT))
	if err != nil {
		return err
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		request := scanner.Text()
		response := handleRequest(request)
		fmt.Fprintf(conn, "%s\n", response)
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading from connection:", err)
	}
}

func handleRequest(request string) string {
	payment := payment.FromString(request)
	response := payment.Process()
	return response.ToString()
}
