package main

import (
	"novachat-server/protocol"

	"golang.org/x/net/websocket"
)

func Rand(n int) []byte {
	b := make([]byte, 0, n)
	for range n {
		b = append(b, byte('s'))
	}
	return b
}

func main() {
	client, err := websocket.Dial("ws://localhost:8080/ws", "", "ws://localhost:123")
	if err != nil {
		panic(err)
	}
	err = protocol.WritePacket(client, protocol.NewPacket(Rand(1000000)))
	if err != nil {
		panic(err)
	}
	err = client.Close()
	if err != nil {
		panic(err)
	}
}
