package mocks

import (
	"github.com/stretchr/testify/mock"
	"net"
	"sync"
)

type MockNetListener struct {
	mock.Mock
}

func (b *MockNetListener) Listen(network string, address string) (net.Listener, error) {
	args := b.Called()
	return args.Get(0).(net.Listener), args.Error(1)
}

type MockListener struct {
	mock.Mock
	wg               sync.WaitGroup
	mu               sync.Mutex
	connectionsCount int
}

func NewMockListener() *MockListener {
	l := &MockListener{wg: sync.WaitGroup{}}
	l.wg.Add(1)
	return l
}

func (m *MockListener) Accept() (net.Conn, error) {
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

func (m *MockListener) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockListener) Addr() net.Addr {
	args := m.Called()
	return args.Get(0).(net.Addr)
}
