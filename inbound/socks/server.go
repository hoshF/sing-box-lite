package socks

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"slices"

	"github.com/hoshF/sing-box-lite/outbound"
	"github.com/hoshF/sing-box-lite/transport"
)

const (
	SOCKS5Version = 0x05
	MethodNoAuth  = 0x00

	CmdConnect = 0x01

	AtypIPv4   = 0x01
	AtypDomain = 0x03
	AtypIPv6   = 0x04

	RepSuccess              = 0x00
	RepGeneralFailure       = 0x01
	RepConnectionNotAllowed = 0x02
	RepNetworkUnreachable   = 0x03
	RepHostUnreachable      = 0x04
	RepConnectionRefused    = 0x05
	RepCommandNotSupported  = 0x07
	RepAddressNotSupported  = 0x08
)

type Request struct {
	Host string
	Port uint16
}

func (r *Request) Address() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

func HandleConnection(conn net.Conn, out outbound.Outbound) error {
	if err := handleHandshake(conn); err != nil {
		return fmt.Errorf("Handshake failed: %w", err)
	}

	fmt.Println("SOCKS5 Handshake success")

	request, err := handleRequest(conn)
	if err != nil {
		return fmt.Errorf("Request parsing failed: %w", err)
	}

	fmt.Printf("destination: %s\n", request.Address())

	targetConn, err := out.Dial(request.Address())
	if err != nil {
		sendReply(conn, RepHostUnreachable)
		return fmt.Errorf("Outbound connect faild: %w", err)
	}
	defer targetConn.Close()

	fmt.Printf("Connected to  %s (through %s)\n", request.Address(), out.Name())

	if err := sendReply(conn, RepSuccess); err != nil {
		return fmt.Errorf("Send reply faild: %w", err)
	}

	fmt.Println("Starting Relay data...")
	if err := transport.Relay(conn, targetConn); err != nil {
		if err != io.EOF {
			return fmt.Errorf("Relay error: %w", err)
		}
	}
	fmt.Println("Connect close")

	return nil
}

func handleHandshake(conn net.Conn) error {
	header := make([]byte, 2)
	if _, err := io.ReadFull(conn, header); err != nil {
		return fmt.Errorf("Read Handshake failed: %w", err)
	}

	version := header[0]
	nMethods := header[1]

	if version != SOCKS5Version {
		return fmt.Errorf("NO support SOCKS5: %d", version)
	}

	methods := make([]byte, nMethods)
	if _, err := io.ReadFull(conn, methods); err != nil {
		return fmt.Errorf("Read auth methods failed: %w", err)
	}

	if !slices.Contains(methods, MethodNoAuth) {
		conn.Write([]byte{SOCKS5Version, 0xFF})
		return errors.New("client does not support no-auth")
	}

	_, err := conn.Write([]byte{SOCKS5Version, MethodNoAuth})
	if err != nil {
		return fmt.Errorf("Send Handshake Respond failed: %w", err)
	}

	return nil
}

func handleRequest(conn net.Conn) (*Request, error) {
	header := make([]byte, 4)
	if _, err := io.ReadFull(conn, header); err != nil {
		return nil, fmt.Errorf("Read requset faild: %w", err)
	}

	version := header[0]
	cmd := header[1]
	atyp := header[3]

	if version != SOCKS5Version {
		return nil, fmt.Errorf("NO support SOCKS5: %d", version)
	}

	if cmd != CmdConnect {
		sendReply(conn, RepCommandNotSupported)
		return nil, fmt.Errorf("NO support cmd: %d (only support CONNECT)", cmd)
	}

	var host string
	switch atyp {
	case AtypIPv4:
		addr := make([]byte, 4)
		if _, err := io.ReadFull(conn, addr); err != nil {
			return nil, fmt.Errorf("Read IPv4 faild: %w", err)
		}
		host = net.IP(addr).String()

	case AtypDomain:
		lenBuf := make([]byte, 1)
		if _, err := io.ReadFull(conn, lenBuf); err != nil {
			return nil, fmt.Errorf("Read domain length faild: %w", err)
		}
		domainLen := lenBuf[0]

		domain := make([]byte, domainLen)
		if _, err := io.ReadFull(conn, domain); err != nil {
			return nil, fmt.Errorf("Read domain faild: %w", err)
		}
		host = string(domain)

	case AtypIPv6:
		addr := make([]byte, 16)
		if _, err := io.ReadFull(conn, addr); err != nil {
			return nil, fmt.Errorf("Read IPv6 faild: %w", err)
		}
		host = net.IP(addr).String()

	default:
		sendReply(conn, RepAddressNotSupported)
		return nil, fmt.Errorf("NO support address type: %d", atyp)
	}

	portBuf := make([]byte, 2)
	if _, err := io.ReadFull(conn, portBuf); err != nil {
		return nil, fmt.Errorf("Read port faild: %w", err)
	}
	port := binary.BigEndian.Uint16(portBuf)

	return &Request{
		Host: host,
		Port: port,
	}, nil
}

func sendReply(conn net.Conn, rep byte) error {
	reply := []byte{
		SOCKS5Version,
		rep,
		0x00,
		AtypIPv4,
		0, 0, 0, 0,
		0, 0,
	}
	_, err := conn.Write(reply)
	return err
}
