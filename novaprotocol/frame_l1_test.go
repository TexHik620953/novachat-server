package novaprotocol_test

import (
	"fmt"
	"novachat-server/novaprotocol"
	"testing"

	"github.com/google/uuid"
)

func TestL1BuildParse(t *testing.T) {
	key := byte(10)
	cryptFunc := func(a []byte) ([]byte, error) {
		v := make([]byte, len(a))
		for i, b := range a {
			v[i] = b ^ key
		}
		return v, nil
	}

	frame := novaprotocol.NewL1Frame(novaprotocol.L0FlagIsEncrypted, uuid.Max, []byte("hello world"))
	frame.SetOrigin(uuid.Max)

	data, err := frame.Build(cryptFunc)
	if err != nil {
		t.Error(err)
		return
	}

	frame, err = novaprotocol.ParseL1Frame(data, cryptFunc)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(string(frame.GetData()))

}
