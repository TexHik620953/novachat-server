package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"io"
	"time"

	"golang.org/x/net/websocket"
)

const (
	HEADER_SIZE  = 4
	POSTFIX_SIZE = 4

	BUFFER_SIZE        = 5
	MESSAGE_SIZE_LIMIT = 10 * 1024 * 1024
)

type Packet struct {
	PacketSize uint32
	Data       []byte
}

func NewPacket(data []byte) *Packet {
	p := &Packet{
		Data:       data,
		PacketSize: uint32(len(data) + HEADER_SIZE + POSTFIX_SIZE),
	}
	return p
}
func (p *Packet) Build() []byte {
	b := make([]byte, 0, p.PacketSize)
	b = append(
		binary.LittleEndian.AppendUint32(b, p.PacketSize),
		p.Data...,
	)
	b = binary.LittleEndian.AppendUint32(b, HashSum(b))
	return b
}

func HashSum(data []byte) uint32 {
	h := fnv.New32a()
	h.Write(data)
	return h.Sum32()
}

func ReadPacket(r io.Reader) (*Packet, error) {
	buf := make([]byte, BUFFER_SIZE)
	var completeMessage bytes.Buffer

	firstPacket := true
	totalRead := 0
	targetRead := uint32(0)

	for {
		n, err := r.Read(buf)
		if err != nil {
			if err == io.EOF && totalRead > 0 {
				return nil, io.EOF
			}
			return nil, err
		}

		totalRead += n
		completeMessage.Write(buf[:n])

		if firstPacket {
			firstPacket = false
			targetRead = binary.LittleEndian.Uint32(buf[:4])
		}

		if n < len(buf) || n >= int(targetRead) {
			break
		}

		if totalRead > MESSAGE_SIZE_LIMIT {
			return nil, fmt.Errorf("message limit reached: %d / %d", totalRead, targetRead)
		}
	}

	data := completeMessage.Bytes()
	if len(data) < HEADER_SIZE+POSTFIX_SIZE+1 {
		return nil, fmt.Errorf("invalid packet: zero length")
	}

	hashSum := binary.LittleEndian.Uint32(data[len(data)-POSTFIX_SIZE:])

	if HashSum(data[:len(data)-POSTFIX_SIZE]) != uint32(hashSum) {
		return nil, fmt.Errorf("invalid hashsum")
	}
	p := &Packet{
		Data:       data[HEADER_SIZE : len(data)-POSTFIX_SIZE],
		PacketSize: targetRead,
	}

	return p, nil
}

func WritePacket(conn *websocket.Conn, msg *Packet) error {
	data := msg.Build()
	totalWritten := 0
	for totalWritten < len(data) {
		n, err := conn.Write(data[totalWritten:])
		if err != nil {
			return fmt.Errorf("write error after %d bytes: %w", totalWritten, err)
		}

		totalWritten += n

		if totalWritten < len(data) {
			time.Sleep(10 * time.Millisecond)
		}
	}

	return nil
}
