package handler

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"

	"github.com/DaronWasTaken/go-tcp-chat/v2/server/helper"
	"github.com/DaronWasTaken/go-tcp-chat/v2/server/types"
)

var (
	clients = make([]*types.Client, 0)
	mutex   = &sync.RWMutex{}
)

func SetupClient(conn net.Conn) {
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
	client := &types.Client{
		Username:      username,
		Conn:          conn,
		InboundBuffer: make(chan string),
	}
	go handleClient(client)
}

func handleClient(client *types.Client) {
	go processInbound(client)

	client.InboundBuffer <- fmt.Sprintf("[INFO] Connected as %s", client.Username)
	enterChatMsg := fmt.Sprintf("%s has entered the chat", client.Username)
	broadcastToRoom(enterChatMsg, client)

	mutex.Lock()
	clients = append(clients, client)
	mutex.Unlock()

	go handleClientInput(client)
}

func handleClientInput(client *types.Client) {
	reader := bufio.NewReader(client.Conn)
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				log.Printf("Connection closed by server: %s", client.Conn.RemoteAddr().String())
				mutex.Lock()
				clients = helper.SliceRemove(clients, client)
				mutex.Unlock()
				broadcastToRoom(fmt.Sprintf("%s has left the chat", client.Username), client)
			} else {
				log.Printf("Error reading from server: %s", err)
			}
			break
		}
		msg = strings.TrimSuffix(msg, "\n")
		if msg == "!info" {
			showInfo(client)
		} else {
			msg = fmt.Sprintf("%s: %s", client.Username, msg)
			broadcastToRoom(msg, client)
		}
	}
}

func showInfo(client *types.Client) {
	var buffer bytes.Buffer
	for _, cl := range clients {
		buffer.WriteString(cl.Username)
		buffer.WriteString(", ")
	}
	msg := buffer.String()
	if msg != "" {
		msg = msg[:len(msg)-2]
	}
	fmt.Fprintf(client.Conn, "[INFO] Users in room: %s\n", msg)
}

func broadcastToRoom(msg string, sender *types.Client) {
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

func processInbound(client *types.Client) {
	for msg := range client.InboundBuffer {
		fmt.Fprintf(client.Conn, "%s\n", msg)
	}
}
