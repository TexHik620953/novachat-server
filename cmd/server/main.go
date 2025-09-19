package main

import (
	"fmt"
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
	Name string
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

		if msg.GetTarget() == protocol.Servercast {
			// For server, do not resend
			err := processServerMessage(client, msg)
			if err != nil {
				log.Printf("failed to process server message: %s", err.Error())
				continue
			}
		} else if msg.GetTarget() == protocol.Broadcast {
			// Broadcast
			msg.SetOrigin(client.ID)
			broadcastMessage(msg)
		} else {
			msg.SetOrigin(client.ID)
			// Unicast
			target, ok := clients.Get(msg.GetTarget())
			if !ok {
				log.Printf("target not found")
				continue
			}
			if err := protocol.WritePacket(target.Conn, msg); err != nil {
				log.Printf("failed to send message to client [%s]: %s", target.ID.String(), err.Error())
				continue
			}
		}

	}
}

func broadcastMessage(p protocol.Packet) {
	clients.Foreach(func(u uuid.UUID, c *Client) {
		if err := protocol.WritePacket(c.Conn, p); err != nil {
			log.Printf("failed to send message to client [%s]: %s", c.ID.String(), err.Error())
		}
	})
}

func processServerMessage(client *Client, msg protocol.Packet) error {
	msgType, err := protocol.ParseMessageType(msg.GetData())
	if err != nil {
		return fmt.Errorf("failed to get message type: %w", err)
	}

	switch msgType {
	case "pr_req":
		request, err := protocol.ParsePresenseRequest(msg.GetData())
		if err != nil {
			return fmt.Errorf("failed to unmarshal: %w", err)
		}
		client.Name = request.Name

		// Send back his uuid
		response, err := protocol.NewPresenseResponse(&protocol.PresenseResponse{
			UserID: client.ID,
		})
		if err != nil {
			return fmt.Errorf("failed to create response: %w", err)
		}
		err = protocol.WritePacket(client.Conn, protocol.NewPacket(uuid.Nil, response, protocol.PacketParams{}))
		if err != nil {
			return fmt.Errorf("failed to write packet: %w", err)
		}
		// Broadcast new user
		br, err := protocol.NewPresenseInfo(&protocol.PresenseInfo{
			UserID: client.ID,
			Name:   client.Name,
		})
		if err != nil {
			return fmt.Errorf("failed to create response: %w", err)
		}
		broadcastMessage(protocol.NewPacket(uuid.Nil, br, protocol.PacketParams{}))
	}

	return nil
}
