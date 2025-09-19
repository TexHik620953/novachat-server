package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"io"
	"time"

	"github.com/google/uuid"
	"golang.org/x/net/websocket"
)

const (
	HEADER_SIZE  = 1 + 4 + 16 + 16
	POSTFIX_SIZE = 4

	BUFFER_SIZE        = 128
	MESSAGE_SIZE_LIMIT = 10 * 1024 * 1024
)
const (
	FlagNone byte = 1 << iota
	FlagEncrypted
	FlagFile
)

var (
	Servercast = uuid.Nil
	Broadcast  = uuid.Max
)

type Packet interface {
	IsEncrypted() bool
	IsFile() bool
	SetData(data []byte)
	GetData() []byte

	SetOrigin(origin uuid.UUID)
	GetOrigin() uuid.UUID

	GetTarget() uuid.UUID

	Build() []byte
}
type packetImpl struct {
	packetSize uint32
	flags      byte
	origin     uuid.UUID
	target     uuid.UUID
	Data       []byte
}

func packetFromBytes(data []byte) (Packet, error) {
	if len(data) < HEADER_SIZE+POSTFIX_SIZE+1 {
		return nil, fmt.Errorf("invalid packet: zero length")
	}

	hashSum := binary.LittleEndian.Uint32(data[len(data)-POSTFIX_SIZE:])

	if HashSum(data[:len(data)-POSTFIX_SIZE]) != uint32(hashSum) {
		return nil, fmt.Errorf("invalid hashsum")
	}
	p := &packetImpl{
		Data:       data[HEADER_SIZE : len(data)-POSTFIX_SIZE],
		packetSize: binary.LittleEndian.Uint32(data[:4]),
		flags:      data[4],
		origin:     uuid.UUID(data[5 : 5+16]),
		target:     uuid.UUID(data[5+16 : 5+16+16]),
	}
	return p, nil
}

type PacketParams struct {
	IsFile       bool
	IsEncrytpted bool
}

func NewPacket(target uuid.UUID, data []byte, params PacketParams) Packet {
	p := &packetImpl{
		flags:      0,
		Data:       data,
		packetSize: uint32(len(data) + HEADER_SIZE + POSTFIX_SIZE),
		target:     target,
	}
	if params.IsEncrytpted {
		p.flags |= FlagEncrypted
	}
	if params.IsFile {
		p.flags |= FlagFile
	}
	return p
}

func (p *packetImpl) IsEncrypted() bool {
	return p.flags&FlagEncrypted != 0
}
func (p *packetImpl) IsFile() bool {
	return p.flags&FlagFile != 0
}
func (p *packetImpl) SetData(data []byte) {
	p.Data = data
}
func (p *packetImpl) SetOrigin(origin uuid.UUID) {
	p.origin = origin
}
func (p *packetImpl) GetOrigin() uuid.UUID {
	return p.origin
}
func (p *packetImpl) GetTarget() uuid.UUID {
	return p.target
}
func (p *packetImpl) GetData() []byte {
	return p.Data
}

func (p *packetImpl) Build() []byte {
	b := make([]byte, 0, p.packetSize)
	b = binary.LittleEndian.AppendUint32(b, p.packetSize)
	b = append(b, p.flags)
	b = append(b, p.origin[:]...)
	b = append(b, p.target[:]...)
	b = append(b, p.Data...)
	b = binary.LittleEndian.AppendUint32(b, HashSum(b))
	return b
}

func HashSum(data []byte) uint32 {
	h := fnv.New32a()
	h.Write(data)
	return h.Sum32()
}

func ReadPacket(r io.Reader) (Packet, error) {
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
			if len(buf) < HEADER_SIZE {
				return nil, fmt.Errorf("no header")
			}
			firstPacket = false
			targetRead = binary.LittleEndian.Uint32(buf[:4])
		}

		if uint32(totalRead) >= targetRead {
			break
		}

		if totalRead > MESSAGE_SIZE_LIMIT {
			return nil, fmt.Errorf("message limit reached: %d / %d", totalRead, targetRead)
		}
	}

	return packetFromBytes(completeMessage.Bytes())
}

func WritePacket(conn *websocket.Conn, msg Packet) error {
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
