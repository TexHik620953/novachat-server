package main

import (
	"novachat-server/protocol"

	"golang.org/x/net/websocket"
)

func main() {
	client, err := websocket.Dial("ws://localhost:8080/ws", "", "ws://localhost:123")
	if err != nil {
		panic(err)
	}
	err = protocol.WritePacket(client, protocol.NewPacket([]byte("Aboba")))
	if err != nil {
		panic(err)
	}
}
