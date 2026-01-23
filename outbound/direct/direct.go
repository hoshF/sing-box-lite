package direct

import (
	"fmt"
	"net"
	"time"
)

type Direct struct {
	timeout time.Duration
}

func New() *Direct {
	return &Direct{
		timeout: 10 * time.Second,
	}
}

func (d *Direct) Name() string {
	return "direct"
}

func (d *Direct) Dial(address string) (net.Conn, error) {
	conn, err := net.DialTimeout("tcp", address, d.timeout)
	if err != nil {
		return nil, fmt.Errorf("Connect %s faild: %w", address, err)
	}
	return conn, nil
}
