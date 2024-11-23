package main

import (
	"fmt"
	"log"
	"net"

	"github.com/DaronWasTaken/go-tcp-chat/server/handler"
)

var (
	port = 8080
)

func main() {
	sock, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	log.Printf("Started listening on port: %d\n", port)
	if err != nil {
		log.Fatal(err)
	}
	defer sock.Close()

	for {
		conn, err := sock.Accept()
		if err != nil {
			log.Printf("Error accepting connection from %s: %s\n", conn.RemoteAddr(), err)
		}
		log.Printf("Client connected: %s", conn.RemoteAddr())
		go handler.SetupClient(conn)
	}
}
