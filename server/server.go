package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
)

const (
	CONN_TYPE   = "tcp"
	CONN_IP     = "192.168.0.44:3000"
	INTERNAL_IP = "192.168.0.44:3001"
)

var (
	CONN_IP2 string
)

func main() {

	ip2 := flag.String("ip", "0.0.0.0:3000", "The other ip that you want to chat with")
	flag.Parse()
	CONN_IP2 = *ip2

	var wg sync.WaitGroup

	wg.Add(2)
	fmt.Printf("Server listening on: %s\n", CONN_IP)

	go func() {
		defer wg.Done()
		startServer(CONN_IP)
	}()

	go func() {
		defer wg.Done()
		startServer(INTERNAL_IP)
	}()

	wg.Wait()
}

func startServer(ip string) {
	listener, err := net.Listen(CONN_TYPE, ip)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	defer listener.Close()
	conn, err := listener.Accept()
	if err != nil {
		log.Fatalf("Error accepting connection: %s", err)
	}
	log.Printf("Client connected: %s\n", conn.RemoteAddr().String())
	defer conn.Close()
	readFromServer(conn)
}

func readFromServer(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				log.Printf("Connection closed by server: %s", conn.RemoteAddr().String())
			} else {
				log.Printf("Error reading from server: %s", err)
			}
			break
		}
		fmt.Print(message)
	}
}
