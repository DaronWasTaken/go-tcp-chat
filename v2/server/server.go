package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"
)

var (
	clients = make([]*Client, 0)
	mutex   = &sync.RWMutex{}
)

func main() {
	sock, err := net.Listen("tcp", "localhost:8080")
	log.Printf("Started listening on port: %d\n", 8080)
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
		go setupClient(conn)
	}
}

func setupClient(conn net.Conn) {
	reader := bufio.NewReader(conn)
		msg, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				log.Printf("Connection closed by server: %s", conn.RemoteAddr().String())
			} else {
				log.Printf("Error reading from server: %s", err)
			}
			return
		}
		username := strings.TrimSuffix(msg, "\n")
		client := &Client{username, conn, make(chan string)}
		go handleClient(client)
}

func handleClient(client *Client) {
	go processInbound(client)
	client.InboundBuffer <- fmt.Sprintf("[INFO] Connected as %s", client.Username)

	enterChatMsg := fmt.Sprintf("%s has entered the chat", client.Username)
	broadcastToRoom(enterChatMsg, client)

	mutex.Lock()
	clients = append(clients, client)
	index := len(clients) - 1
	mutex.Unlock()

	reader := bufio.NewReader(client.Conn)
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				log.Printf("Connection closed by server: %s", client.Conn.RemoteAddr().String())
				mutex.Lock()
				clients = sliceRemove(clients, index)
				mutex.Unlock()
				broadcastToRoom(fmt.Sprintf("%s has left the chat", client.Username), client)
			} else {
				log.Printf("Error reading from server: %s", err)
			}
			break
		}
		msg = strings.TrimSuffix(msg, "\n")
		if (msg == "!info") {
			showInfo(client)
		} else {
			msg = fmt.Sprintf("%s: %s", client.Username, msg)
			broadcastToRoom(msg, client)
		}
	}
}

func showInfo(client *Client) {
	var buffer bytes.Buffer
	for _, cl := range clients {
		buffer.WriteString(cl.Username)
		buffer.WriteString(", ")
	}
	msg := buffer.String()
	if msg != "" {
		msg = msg[:len(msg)-2]
	}
	fmt.Fprintf(client.Conn, "[INFO] Users in room: %s\n",msg)
}

func broadcastToRoom(msg string, sender *Client) {
	log.Printf("Broadcasting message from %q: %q\n", sender.Conn.RemoteAddr(), msg)
	mutex.RLock()
	defer mutex.RUnlock()
	for _, cl := range clients {
		select {
		case cl.InboundBuffer <- msg:
		default:
			log.Printf("%s: inbound channel full", cl.Username)
		}
	}
}

func processInbound(client *Client) {
	for msg := range client.InboundBuffer {
		fmt.Fprintf(client.Conn, "%s\n", msg)
	}
}

func sliceRemove[T any](slice []T, index int) []T {
	slice[index] = slice[len(slice)-1]
	return slice[:len(slice)-1]
}

type Client struct {
	Username      string
	Conn          net.Conn
	InboundBuffer chan string
}
