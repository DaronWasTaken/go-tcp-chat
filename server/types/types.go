package types

import "net"

type Client struct {
	Username      string
	Conn          net.Conn
	InboundBuffer chan string
}