package protocol

import (
	"encoding/json"
	"math/rand/v2"
	"strings"
)

const randRunes = "ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890-=_./!@#$%^&*():;'[]<>?"

type messageImpl[T any] struct {
	Data T      `json:"data"`
	Type string `json:"string"`
	Salt string `json:"salt"`
}

func ParseMessageType(msg []byte) (string, error) {
	var r messageImpl[interface{}]
	if err := json.Unmarshal(msg, &r); err != nil {
		return "", err
	}
	return r.Type, nil
}

func newMessage[T any](msgType string, data T) ([]byte, error) {
	var sb strings.Builder
	for _, k := range rand.Perm(len(randRunes)) {
		sb.WriteByte(randRunes[k])
	}
	m := &messageImpl[T]{
		Salt: sb.String(),
		Data: data,
		Type: msgType,
	}
	return json.Marshal(m)
}
