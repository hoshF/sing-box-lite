package main

import (
	"fmt"
	"log"
	"net"

	"github.com/hoshF/sing-box-lite/inbound/socks"
)

func main() {
	listener, err := net.Listen("tcp", ":1080")
	if err != nil {
		log.Fatal("listen fail:", err)
	}
	defer listener.Close()
	fmt.Println("SOCKS5 proxy started, listening :1080")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("accept connection fail", err)
			continue
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	clientAddr := conn.RemoteAddr().String()
	fmt.Printf("new connection: %s\n", clientAddr)

	if err := socks.HandleConnection(conn); err != nil {
		fmt.Printf("Handle fail: %v\n", err)
	}
}
