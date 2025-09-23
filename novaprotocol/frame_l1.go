package novaprotocol

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"

	"github.com/google/uuid"
)

const (
	L1FlagIsJson byte = 1 << iota
	L1FlagIsFile
	L1FlagIsEncrypted
)

const (
	l1HeaderSize        = 1 + 16 + 16          // flags + origin + destination
	l1MinFrameSize      = l1HeaderSize + 8 + 4 // header + min data (salt) + CRC
	l1flagsOffset       = 0
	l1originOffset      = 1
	l1destinationOffset = 17
	l1dataOffset        = 33
)

// NovaFrameL1 represents Layer 1 frame structure
type NovaFrameL1 struct {
	flags       byte
	origin      uuid.UUID
	destination uuid.UUID
	data        []byte
}

// NewL1Frame creates a new L1 frame
func NewL1Frame(flags byte, destination uuid.UUID, data []byte) *NovaFrameL1 {
	return &NovaFrameL1{
		flags:       flags,
		destination: destination,
		data:        data,
	}
}

// GetFlags returns frame flags
func (f *NovaFrameL1) GetFlags() byte {
	return f.flags
}

// SetFlags sets frame flags
func (f *NovaFrameL1) SetFlags(flags byte) {
	f.flags = flags
}

// GetData returns frame data
func (f *NovaFrameL1) GetData() []byte {
	return f.data
}

// GetDestination returns destination UUID
func (f *NovaFrameL1) GetDestination() uuid.UUID {
	return f.destination
}

// GetOrigin returns origin UUID
func (f *NovaFrameL1) GetOrigin() uuid.UUID {
	return f.origin
}

// SetOrigin sets origin UUID
func (f *NovaFrameL1) SetOrigin(o uuid.UUID) {
	f.origin = o
}

// Build constructs the L1 frame bytes with optional encryption
func (f *NovaFrameL1) Build(encryptFunc CryptFunc) ([]byte, error) {
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
	if f.flags&L1FlagIsEncrypted != 0 && encryptFunc == nil {
		return nil, fmt.Errorf("encryption required but no encrypt function provided")
	}

	// Encrypt if needed
	var processedData []byte
	var err error

	if f.flags&L1FlagIsEncrypted != 0 {
		processedData, err = encryptFunc(content)
		if err != nil {
			return nil, fmt.Errorf("encryption failed: %w", err)
		}
	} else {
		processedData = content
	}

	// Create frame buffer
	frameSize := l1HeaderSize + len(processedData) + crcSize
	frameData := make([]byte, 0, frameSize)

	// Build header
	frameData = append(frameData, f.flags)
	frameData = append(frameData, f.origin[:]...)
	frameData = append(frameData, f.destination[:]...)
	frameData = append(frameData, processedData...)

	// Calculate CRC
	crc, err := calculateCRC(frameData[:l1HeaderSize], content)
	if err != nil {
		return nil, fmt.Errorf("CRC calculation failed: %w", err)
	}

	// Append CRC
	frameData = binary.LittleEndian.AppendUint32(frameData, crc)

	return frameData, nil
}

// ParseL1Frame parses raw bytes into NovaFrameL1
func ParseL1Frame(data []byte, decryptFunc CryptFunc) (*NovaFrameL1, error) {
	// Validate minimum length
	if len(data) < l1MinFrameSize {
		return nil, ErrorFrameZeroLength
	}

	// Extract components
	flags := data[l1flagsOffset]
	originBytes := data[l1originOffset:l1destinationOffset]
	destinationBytes := data[l1destinationOffset:l1dataOffset]
	encryptedData := data[l1dataOffset : len(data)-crcSize]
	receivedCRC := binary.LittleEndian.Uint32(data[len(data)-crcSize:])

	// Parse UUIDs
	originUUID, err := uuid.FromBytes(originBytes)
	if err != nil {
		return nil, fmt.Errorf("invalid origin UUID: %w", err)
	}

	destinationUUID, err := uuid.FromBytes(destinationBytes)
	if err != nil {
		return nil, fmt.Errorf("invalid destination UUID: %w", err)
	}

	// Validate decryption requirements
	if flags&L1FlagIsEncrypted != 0 && decryptFunc == nil {
		return nil, fmt.Errorf("encrypted frame requires decrypt function")
	}

	// Decrypt if needed
	var decryptedContent []byte

	if flags&L1FlagIsEncrypted != 0 {
		decryptedContent, err = decryptFunc(encryptedData)
		if err != nil {
			return nil, fmt.Errorf("decryption failed: %w", err)
		}
	} else {
		decryptedContent = encryptedData
	}

	// Validate CRC
	header := data[:l1HeaderSize]
	expectedCRC, err := calculateCRC(header, decryptedContent)
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

	return &NovaFrameL1{
		flags:       flags,
		origin:      originUUID,
		destination: destinationUUID,
		data:        frameData,
	}, nil
}
