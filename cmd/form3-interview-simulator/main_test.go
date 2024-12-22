package main

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"net"
	"os"
	"syscall"
	"testing"
	"time"
)

func Test_TestSigIntStopsTheService(t *testing.T) {
	go main()

	time.Sleep(1 * time.Second)

	conn, err := net.Dial("tcp", fmt.Sprintf(":%d", PORT))
	require.NoError(t, err, "Failed to connect to server")
	require.NotNil(t, conn)
	conn.Close()

	p, err := os.FindProcess(os.Getpid())
	if err != nil {
		panic(err)
	}
	p.Signal(syscall.SIGINT)

	time.Sleep(1 * time.Second)

	conn, err = net.Dial("tcp", fmt.Sprintf(":%d", PORT))
	require.Error(t, err)
	require.Nil(t, conn)
}
