package main

import (
	"fmt"
	"log"
	"math/big"
	"novachat-server/common/safemap"
	"novachat-server/protocol"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/google/uuid"
	"github.com/rivo/tview"
	"golang.org/x/net/websocket"
)

var sendChan = make(chan protocol.Packet, 10)
var recvChan = make(chan protocol.Packet, 10)

var name = ""
var userId uuid.UUID

var app = tview.NewApplication()
var header, chatView, logsView *tview.TextView
var inputField *tview.InputField

var priv, pub *big.Int

type UserInfo struct {
	Key  []byte
	Name string
}

var usersInfo = safemap.New[uuid.UUID, *UserInfo]()

func main() {
	// Generate private key
	var err error
	priv, pub, err = protocol.GenerateKeyPair()
	if err != nil {
		panic(fmt.Errorf("failed to generate keys pair: %w", err))
	}

	client, err := websocket.Dial("ws://localhost:8080/ws", "", "ws://localhost:123")
	if err != nil {
		panic(err)
	}
	defer client.Close()

	fmt.Printf("Enter your name: ")
	fmt.Scanln(&name)

	// Send presense
	{
		buf, err := protocol.NewPresenseRequest(&protocol.PresenseRequest{
			Name: name,
		})
		if err != nil {
			panic("failed to make presense request")
		}
		sendChan <- protocol.NewPacket(protocol.Servercast, buf, protocol.PacketParams{})
	}

	go runClient(client)
	go packetsHandler(client)
	runApp()
}

func runApp() {
	// Создаем элементы интерфейса

	header = tview.NewTextView().
		SetTextAlign(tview.AlignLeft).
		SetRegions(false).
		SetWordWrap(true).
		SetDynamicColors(true).
		SetMaxLines(1)

	header.SetBackgroundColor(tcell.ColorSilver)
	header.SetText(fmt.Sprintf("[black]%s", name))

	logsView = tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(false).
		SetWordWrap(true)

	chatView = tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(false).
		SetWordWrap(true)

	inputField = tview.NewInputField().
		SetLabel("$: ").
		SetFieldWidth(0)

	// Layout
	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(header, 1, 1, false).
		AddItem(
			tview.NewFlex().
				SetDirection(tview.FlexColumn).
				AddItem(chatView, 0, 4, false).
				AddItem(logsView, 0, 1, false), 0, 1, false).
		AddItem(inputField, 1, 1, true)

	// Обработка ввода
	inputField.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			defer inputField.SetText("")
			text := inputField.GetText()
			if text != "" {
				go func() {
					msg, err := protocol.NewChatMessage(&protocol.ChatMessage{
						Text: text,
					})
					if err != nil {
						app.QueueUpdateDraw(func() {
							fmt.Fprintf(logsView, "[red]failed to create message: %s\n", err.Error())
							logsView.ScrollToEnd()
						})
					}
					usersInfo.Foreach(func(u uuid.UUID, ui *UserInfo) {
						uMsg, err := protocol.EncryptAES256(ui.Key, msg)
						if err != nil {
							app.QueueUpdateDraw(func() {
								fmt.Fprintf(logsView, "[red]failed to encrypt message for user: %s\n", err.Error())
								logsView.ScrollToEnd()
							})
							return
						}
						sendChan <- protocol.NewPacket(u, uMsg, protocol.PacketParams{IsEncrytpted: true})
					})
					app.QueueUpdateDraw(func() {
						fmt.Fprintf(chatView, "[yellow][LOCAL[][green][%s[][white]: %s\n", name, text)
						chatView.ScrollToEnd()
					})
				}()
			}
		}
	})
	if err := app.SetRoot(flex, true).Run(); err != nil {
		panic(err)
	}
}

func runClient(conn *websocket.Conn) {
	// Send packets
	go func() {
		for packet := range sendChan {
			protocol.WritePacket(conn, packet)
		}
	}()

	// Recv packets
	go func() {
		for {
			packet, err := protocol.ReadPacket(conn)
			if err != nil {
				<-time.After(time.Millisecond * 50)
				continue
			}
			recvChan <- packet
		}
	}()
}
func packetsHandler(conn *websocket.Conn) {
	for recvPacket := range recvChan {

		data := recvPacket.GetData()
		if recvPacket.IsEncrypted() {
			userInfo, ok := usersInfo.Get(recvPacket.GetOrigin())
			if !ok {
				log.Printf("failed to decrypt message from %s, no key\n", recvPacket.GetOrigin())
				app.QueueUpdateDraw(func() {
					fmt.Fprintf(logsView, "[red]failed to decrypt message from %s, no key\n", recvPacket.GetOrigin())
					logsView.ScrollToEnd()
				})
				continue
			}
			decryptedData, err := protocol.DecryptAES256(userInfo.Key, data)
			if err != nil {
				app.QueueUpdateDraw(func() {
					fmt.Fprintf(logsView, "[red]failed to decrypt message: %s\n", err.Error())
					logsView.ScrollToEnd()
				})
				continue
			}
			data = decryptedData
		}

		msgType, err := protocol.ParseMessageType(data)
		if err != nil {
			app.QueueUpdateDraw(func() {
				fmt.Fprintf(logsView, "[red]failed to parse message type: %s\n", err.Error())
				logsView.ScrollToEnd()
			})
			continue
		}

		if recvPacket.GetOrigin() == uuid.Nil {
			// From server
			switch msgType {
			case "pr_res":
				msg, err := protocol.ParsePresenseResponse(data)
				if err != nil {
					app.QueueUpdateDraw(func() {
						fmt.Fprintf(logsView, "[red]failed to parse message: %s\n", err.Error())
						logsView.ScrollToEnd()
					})
					continue
				}
				userId = msg.UserID

				app.QueueUpdateDraw(func() {
					fmt.Fprintf(logsView, "[red][SERVER[][white] Your USERID: [green]%s\n", userId.String())
					logsView.ScrollToEnd()
				})
				// send public key to everyone
				{
					buf, err := protocol.NewPublicKeyMessage(&protocol.PublicKeyMessage{
						PublicKey: pub.Text(62),
					})
					if err != nil {
						app.QueueUpdateDraw(func() {
							fmt.Fprintf(logsView, "[red]failed to make message: %s\n", err.Error())
							logsView.ScrollToEnd()
						})
					}
					sendChan <- protocol.NewPacket(protocol.Broadcast, buf, protocol.PacketParams{})
				}

			case "pr_inf":
				msg, err := protocol.ParsePresenseInfo(data)
				if err != nil {
					app.QueueUpdateDraw(func() {
						fmt.Fprintf(logsView, "[red]failed to parse message: %s\n", err.Error())
						logsView.ScrollToEnd()
					})
					continue
				}
				app.QueueUpdateDraw(func() {
					fmt.Fprintf(logsView, "[red][SERVER[][white] USER [green][%s[] [red]%s[white] joined chat\n", msg.UserID.String(), msg.Name)
					logsView.ScrollToEnd()
				})
				if msg.UserID != userId {
					usersInfo.Set(msg.UserID, &UserInfo{
						Key:  nil,
						Name: msg.Name,
					})
				}
			}
		} else if recvPacket.GetOrigin() != userId {
			// From another user
			switch msgType {
			case "pub":
				msg, err := protocol.ParsePublicKeyMessage(data)
				if err != nil {
					app.QueueUpdateDraw(func() {
						fmt.Fprintf(logsView, "[red]failed to parse message: %s\n", err.Error())
						logsView.ScrollToEnd()
					})
					continue
				}
				targetPub, ok := new(big.Int).SetString(msg.PublicKey, 62)
				if !ok {
					app.QueueUpdateDraw(func() {
						fmt.Fprintf(logsView, "[red]failed to decode public key: %s\n", err.Error())
						logsView.ScrollToEnd()
					})
					continue
				}
				key := protocol.ComputeSharedKey(priv, targetPub)

				userInfo, ok := usersInfo.Get(recvPacket.GetOrigin())
				if !ok {
					userInfo = &UserInfo{}
					usersInfo.Set(recvPacket.GetOrigin(), userInfo)
				}
				// Exchange only once
				if userInfo.Key == nil {
					userInfo.Key = key

					app.QueueUpdateDraw(func() {
						fmt.Fprintf(logsView, "[red][DEBUG[][white] Exchanged keys with: [green]%s\n", recvPacket.GetOrigin().String())
						logsView.ScrollToEnd()
					})
					// Send your key in return
					resp, err := protocol.NewPublicKeyMessage(&protocol.PublicKeyMessage{
						PublicKey: pub.Text(62),
					})
					if err != nil {
						app.QueueUpdateDraw(func() {
							fmt.Fprintf(logsView, "[red]failed to parse pub message: %s\n", err.Error())
							logsView.ScrollToEnd()
						})
						continue
					}
					sendChan <- protocol.NewPacket(recvPacket.GetOrigin(), resp, protocol.PacketParams{})
				}

			case "ch_msg":
				msg, err := protocol.ParseChatMessage(data)
				if err != nil {
					app.QueueUpdateDraw(func() {
						fmt.Fprintf(logsView, "[red]failed to parse message: %s\n", err.Error())
						logsView.ScrollToEnd()
					})
					continue
				}

				userInfo, ok := usersInfo.Get(recvPacket.GetOrigin())
				if !ok {
					app.QueueUpdateDraw(func() {
						fmt.Fprintf(logsView, "[red]msg from unknown user")
						logsView.ScrollToEnd()
					})
					continue
				}
				app.QueueUpdateDraw(func() {
					fmt.Fprintf(chatView, "[green][%s[][white]: %s\n", userInfo.Name, msg.Text)
					chatView.ScrollToEnd()
				})
			}
		}
	}
}
