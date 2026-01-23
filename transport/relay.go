package transport

import (
	"io"
	"net"
	"sync"
)

func Relay(left, right net.Conn) error {
	var wg sync.WaitGroup
	var leftErr, rightErr error

	wg.Add(1)
	go func() {
		defer wg.Done()
		_, leftErr = io.Copy(right, left)
		if conn, ok := right.(*net.TCPConn); ok {
			conn.CloseWrite()
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		_, rightErr = io.Copy(left, right)
		if conn, ok := left.(*net.TCPConn); ok {
			conn.CloseWrite()
		}
	}()

	wg.Wait()

	if leftErr != nil {
		return leftErr
	}
	return rightErr
}
