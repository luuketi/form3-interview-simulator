package mocks

import (
	"github.com/stretchr/testify/mock"
	"net"
	"time"
)

type MockConnection struct {
	mock.Mock
}

func (c *MockConnection) Read([]byte) (n int, err error) {
	args := c.Called()
	return args.Int(0), args.Error(1)
}

func (c *MockConnection) Write([]byte) (n int, err error) {
	args := c.Called()
	return args.Int(0), args.Error(1)
}

func (c *MockConnection) Close() error {
	args := c.Called()
	return args.Error(0)
}

func (c *MockConnection) LocalAddr() net.Addr {
	args := c.Called()
	return args.Get(0).(net.Addr)
}

func (c *MockConnection) RemoteAddr() net.Addr {
	args := c.Called()
	return args.Get(0).(net.Addr)
}

func (c *MockConnection) SetDeadline(time.Time) error {
	args := c.Called()
	return args.Error(0)
}

func (c *MockConnection) SetReadDeadline(time.Time) error {
	args := c.Called()
	return args.Error(0)
}

func (c *MockConnection) SetWriteDeadline(time.Time) error {
	args := c.Called()
	return args.Error(0)
}
