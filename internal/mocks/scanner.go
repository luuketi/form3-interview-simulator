package mocks

import (
	tcp_listener "github.com/form3tech-oss/interview-simulator/internal/tcp-listener"
	"github.com/stretchr/testify/mock"
	"io"
)

type MockNewScanner struct {
	scanner tcp_listener.Scanner
}

func NewMockNewScanner(scanner tcp_listener.Scanner) *MockNewScanner {
	return &MockNewScanner{scanner: scanner}
}

func (b *MockNewScanner) NewScanner(io.Reader) tcp_listener.Scanner {
	return b.scanner
}

type MockBufioScanner struct {
	mock.Mock
}

func (r *MockBufioScanner) Scan() bool {
	args := r.Called()
	return args.Bool(0)
}

func (r *MockBufioScanner) Text() string {
	args := r.Called()
	return args.String(0)
}

func (r *MockBufioScanner) Err() error {
	args := r.Called()
	return args.Error(0)
}
