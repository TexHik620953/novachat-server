package novaprotocol

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
)

const (
	L0FlagNone byte = 1 << iota
	L0FlagIsEncrypted
)
const (
	l0minFrameSize   = 17 // 4(size) + 1(flags) + 8(min data) + 4(crc)
	l0sizeFieldSize  = 4
	l0flagsFieldSize = 1
	l0headerSize     = l0sizeFieldSize + l0flagsFieldSize
)

// NovaFrameL0 represents Layer 0 frame structure
type NovaFrameL0 struct {
	flags byte
	data  []byte
}

// NewL0Frame creates a new L0 frame
func NewL0Frame(flags byte, data []byte) *NovaFrameL0 {
	return &NovaFrameL0{
		flags: flags,
		data:  data,
	}
}

// GetFlags returns frame flags
func (f *NovaFrameL0) GetFlags() byte {
	return f.flags
}

// SetFlags sets frame flags
func (f *NovaFrameL0) SetFlags(flags byte) {
	f.flags = flags
}

// GetData returns frame data
func (f *NovaFrameL0) GetData() []byte {
	return f.data
}

// Build constructs the frame bytes with optional encryption
func (f *NovaFrameL0) Build(encryptFunc CryptFunc) ([]byte, error) {
	// Generate salt
	salt := make([]byte, saltSize)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	// Prepare content: data + salt
	content := make([]byte, 0, len(f.data)+saltSize)
	content = append(content, f.data...)
	content = append(content, salt...)

	// Validate encryption requirements
	if f.flags&L0FlagIsEncrypted != 0 && encryptFunc == nil {
		return nil, fmt.Errorf("encryption required but no encrypt function provided")
	}

	// Encrypt if needed
	var processedData []byte
	var err error

	if f.flags&L0FlagIsEncrypted != 0 {
		processedData, err = encryptFunc(content)
		if err != nil {
			return nil, fmt.Errorf("encryption failed: %w", err)
		}
	} else {
		processedData = content
	}

	// Calculate packet size and create buffer
	packetSize := l0headerSize + len(processedData) + crcSize
	frameData := make([]byte, 0, packetSize)

	// Build frame
	frameData = binary.LittleEndian.AppendUint32(frameData, uint32(packetSize))
	frameData = append(frameData, f.flags)
	frameData = append(frameData, processedData...)

	// Calculate CRC
	crc, err := calculateCRC(frameData[:l0headerSize], content)
	if err != nil {
		return nil, fmt.Errorf("CRC calculation failed: %w", err)
	}

	// Append CRC
	frameData = binary.LittleEndian.AppendUint32(frameData, crc)

	return frameData, nil
}

// ParseL0Frame parses raw bytes into NovaFrameL0
func ParseL0Frame(data []byte, decryptFunc CryptFunc) (*NovaFrameL0, error) {
	// Validate minimum length
	if len(data) < l0minFrameSize {
		return nil, ErrorFrameZeroLength
	}

	// Parse and validate size field
	declaredSize := binary.LittleEndian.Uint32(data[:l0sizeFieldSize])
	if declaredSize != uint32(len(data)) {
		return nil, ErrorFrameSizeMissmatch
	}

	// Extract components
	flags := data[l0sizeFieldSize]
	encryptedData := data[l0headerSize : len(data)-crcSize]
	receivedCRC := binary.LittleEndian.Uint32(data[len(data)-crcSize:])

	// Validate decryption requirements
	if flags&L0FlagIsEncrypted != 0 && decryptFunc == nil {
		return nil, fmt.Errorf("encrypted frame requires decrypt function")
	}

	// Decrypt if needed
	var decryptedContent []byte
	var err error

	if flags&L0FlagIsEncrypted != 0 {
		decryptedContent, err = decryptFunc(encryptedData)
		if err != nil {
			return nil, fmt.Errorf("decryption failed: %w", err)
		}
	} else {
		decryptedContent = encryptedData
	}

	// Validate CRC
	expectedCRC, err := calculateCRC(data[:l0headerSize], decryptedContent)
	if err != nil {
		return nil, fmt.Errorf("CRC calculation failed: %w", err)
	}

	if expectedCRC != receivedCRC {
		return nil, ErrorFrameInvalidHashSum
	}

	// Extract data (remove salt)
	if len(decryptedContent) < saltSize {
		return nil, fmt.Errorf("frame content too short for salt")
	}

	frameData := decryptedContent[:len(decryptedContent)-saltSize]

	return &NovaFrameL0{
		flags: flags,
		data:  frameData,
	}, nil
}
