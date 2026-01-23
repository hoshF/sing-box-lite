package outbound

import "net"

type Outbound interface {
	Name() string

	Dial(address string) (net.Conn, error)
}
