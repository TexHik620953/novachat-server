package protocol

import "encoding/json"

type ChatMessage struct {
	Text string `json:"text"`
}

func ParseChatMessage(msg []byte) (*ChatMessage, error) {
	var r messageImpl[*ChatMessage]
	if err := json.Unmarshal(msg, &r); err != nil {
		return nil, err
	}
	return r.Data, nil
}
func NewChatMessage(msg *ChatMessage) ([]byte, error) {
	return newMessage("ch_msg", msg)
}
