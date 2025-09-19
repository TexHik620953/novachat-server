package main

import (
	"io"
	"log"
	"net/http"
	"novachat-server/common/safemap"
	"novachat-server/protocol"
	"time"

	"github.com/google/uuid"
	"golang.org/x/net/websocket"
)

// Client представляет подключенного клиента
type Client struct {
	ID   uuid.UUID
	Conn *websocket.Conn
}

var clients = safemap.New[uuid.UUID, *Client]()

func main() {
	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.Handle("/ws", websocket.Handler(handleWebSocket))
	log.Println("server started on port :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleWebSocket(ws *websocket.Conn) {
	id, err := uuid.NewRandom()
	if err != nil {
		return
	}
	client := &Client{
		ID:   id,
		Conn: ws,
	}
	clients.Set(client.ID, client)
	defer clients.Remove(client.ID)

	for {
		msg, err := protocol.ReadPacket(client.Conn)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Printf("failed to read packet: %s", err.Error())
			<-time.After(time.Millisecond * 10)
			break
		}
		log.Printf("received message with data: %s\n", string(msg.Data))
		broadcastMessage(msg)
	}
}

func broadcastMessage(p *protocol.Packet) {
	clients.Foreach(func(u uuid.UUID, c *Client) {
		if err := protocol.WritePacket(c.Conn, p); err != nil {
			log.Printf("failed to send message to client [%s]: %s", c.ID.String(), err.Error())
		}
	})
}
