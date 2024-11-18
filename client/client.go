package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"time"
)

const (
	CONN_TYPE      = "tcp"
	SERVER_IP      = "192.168.0.44:3001"
	RETRY_INTERVAL = 5 * time.Second
	MAX_RETRIES    = 5
)

var (
	CONN_IP2 string
	USERNAME string
)

func main() {
	clear()
	ip2 := flag.String("ip", "0.0.0.0:3000", "The other ip that you want to chat with")
	username := flag.String("u", "default", "Your username")
	flag.Parse()
	CONN_IP2 = *ip2
	USERNAME = *username
	startClient()
}

func startClient() {
	var conn net.Conn
	var err error
	for retries := 0; retries < MAX_RETRIES; retries++ {
		conn, err = net.Dial(CONN_TYPE, CONN_IP2)
		if err == nil {
			log.Println("Connected to server: ", conn.RemoteAddr().String())
			break
		}
		log.Printf("Failed to connect to server: %s. Retrying in %s...", err, RETRY_INTERVAL)
		time.Sleep(RETRY_INTERVAL)
	}
	if err != nil {
		log.Fatalf("Could not connect to server after %d retries: %s", MAX_RETRIES, err)
		os.Exit(1)
	}

	conn2, err := net.Dial(CONN_TYPE, SERVER_IP)
	if err != nil {
		log.Fatalf("Could not connect to server after %d retries: %s", MAX_RETRIES, err)
		os.Exit(1)
	}

	readFromStdinAndSend(conn, conn2)

	defer conn.Close()
	defer conn2.Close()
}

func readFromStdinAndSend(conn net.Conn, conn2 net.Conn) {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		if !scanner.Scan() {
			break
		}
		message := fmt.Sprintf("%s: %s", USERNAME, scanner.Text())
		if _, err := fmt.Fprintln(conn, message); err != nil {
			log.Printf("Error writing to connection: %s", err)
			break
		}
		if _, err := fmt.Fprintln(conn2, message); err != nil {
			log.Printf("Error writing to connection: %s", err)
			break
		}
		clear()
	}
	if err := scanner.Err(); err != nil {
		log.Printf("Error reading from input: %s", err)
	}
}

func clear() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}
