package novaprotocol_test

import (
	"fmt"
	"novachat-server/novaprotocol"
	"testing"

	"github.com/google/uuid"
)

func TestL0BuildParse(t *testing.T) {
	key := byte(10)
	cryptFunc := func(a []byte) ([]byte, error) {
		v := make([]byte, len(a))
		for i, b := range a {
			v[i] = b ^ key
		}
		return v, nil
	}

	frame := novaprotocol.NewL0Frame(novaprotocol.L0FlagIsEncrypted, uuid.Nil, []byte("hello world"))

	data, err := frame.Build(cryptFunc)
	if err != nil {
		t.Error(err)
		return
	}

	frame, err = novaprotocol.ParseL0Frame(data, cryptFunc)
	if err != nil {
		t.Error(err)
		return
	}

	if frame.GetDestination() != uuid.Nil {
		t.Errorf("destination missmatch")
	}
	fmt.Println(string(frame.GetData()))

}
