package novaprotocol_test

import (
	"fmt"
	"novachat-server/novaprotocol"
	"testing"

	"github.com/google/uuid"
)

var (
	l0cryptFunc = func(a []byte) ([]byte, error) {
		key := byte(10)
		v := make([]byte, len(a))
		for i, b := range a {
			v[i] = b ^ key
		}
		return v, nil
	}
	l1cryptFunc = func(a []byte) ([]byte, error) {
		key := byte(0xfa)
		v := make([]byte, len(a))
		for i, b := range a {
			v[i] = b ^ key
		}
		return v, nil
	}
)

func TestProtocol(t *testing.T) {
	fFrameData := novaprotocol.NewFileBlockFrame(0, uuid.Max, []byte("hello world"))

	l1FrameData, err := novaprotocol.NewL1Frame(novaprotocol.L1FlagIsEncrypted|novaprotocol.L1FlagIsFile, uuid.Nil, fFrameData).Build(l1cryptFunc)
	if err != nil {
		t.Error(err)
		return
	}

	l0FrameData, err := novaprotocol.NewL0Frame(novaprotocol.L0FlagIsEncrypted, l1FrameData).Build(l0cryptFunc)
	if err != nil {
		t.Error(err)
		return
	}
	// Convert back

	l0, err := novaprotocol.ParseL0Frame(l0FrameData, l0cryptFunc)
	if err != nil {
		t.Error(err)
		return
	}
	l1, err := novaprotocol.ParseL1Frame(l0.GetData(), l1cryptFunc)
	if err != nil {
		t.Error(err)
		return
	}
	block, err := novaprotocol.ParseFileBlockFrame(l1.GetData())
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(string(block.Data))
}
