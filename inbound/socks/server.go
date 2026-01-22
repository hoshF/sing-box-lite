package socks

import (
	"errors"
	"fmt"
	"io"
	"net"
	"slices"
)

const (
	SOCKS5Version = 0x05
	MethodNoAuth  = 0x00
)

func HandleConnection(conn net.Conn) error {
	if err := handleHandshake(conn); err != nil {
		return fmt.Errorf("Handshake fail: %w", err)
	}

	fmt.Println("SOCKS5 Handshake success")

	return nil
}

func handleHandshake(conn net.Conn) error {
	header := make([]byte, 2)
	if _, err := io.ReadFull(conn, header); err != nil {
		return fmt.Errorf("Read Handshake fail: %w", err)
	}

	version := header[0]
	nMethods := header[1]

	if version != SOCKS5Version {
		return fmt.Errorf("NO support SOCKS5 version: %d", version)
	}

	methods := make([]byte, nMethods)
	if _, err := io.ReadFull(conn, methods); err != nil {
		return fmt.Errorf("Read auth methods fail: %w", err)
	}

	if !slices.Contains(methods, MethodNoAuth) {
		conn.Write([]byte{SOCKS5Version, 0xFF})
		return errors.New("client does not support no-auth")
	}

	_, err := conn.Write([]byte{SOCKS5Version, MethodNoAuth})
	if err != nil {
		return fmt.Errorf("Send Handshake Respond fail: %w", err)
	}

	return nil
}
