package novaprotocol_test

import (
	"fmt"
	"novachat-server/novaprotocol"
	"testing"
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

	frame := novaprotocol.NewL0Frame(novaprotocol.L0FlagIsEncrypted, []byte("hello world"))

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
	fmt.Println(string(frame.GetData()))

}
