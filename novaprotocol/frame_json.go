package novaprotocol

import (
	"encoding/json"
)

type messageImpl[T any] struct {
	Data T      `json:"data"`
	Type string `json:"type"`
}

func ParseJsonMessageType(msg []byte) (string, error) {
	var r struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(msg, &r); err != nil {
		return "", err
	}
	return r.Type, nil
}

func ParseJsonMessage[T any](jsonMsg []byte) (*T, error) {
	var r messageImpl[*T]
	if err := json.Unmarshal(jsonMsg, &r); err != nil {
		return nil, err
	}
	return r.Data, nil
}
func NewJsonMessage[T any](msgType string, data T) ([]byte, error) {
	m := &messageImpl[T]{
		Data: data,
		Type: msgType,
	}
	return json.Marshal(m)
}
