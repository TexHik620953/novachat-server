package novaprotocol

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
)

const (
	L1FlagIsJson byte = 1 << iota
	L1FlagIsFile
	L1FlagIsEncrypted
)

const (
	l1HeaderSize   = 1                    // flags
	l1MinFrameSize = l1HeaderSize + 8 + 4 // header + min data (salt) + CRC
	l1flagsOffset  = 0
	l1dataOffset   = 1
)

// NovaFrameL1 represents Layer 1 frame structure
type NovaFrameL1 struct {
	flags byte
	data  []byte
}

// NewL1Frame creates a new L1 frame
func NewL1Frame(flags byte, data []byte) *NovaFrameL1 {
	return &NovaFrameL1{
		flags: flags,
		data:  data,
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
	encryptedData := data[l1dataOffset : len(data)-crcSize]
	receivedCRC := binary.LittleEndian.Uint32(data[len(data)-crcSize:])

	// Validate decryption requirements
	if flags&L1FlagIsEncrypted != 0 && decryptFunc == nil {
		return nil, fmt.Errorf("encrypted frame requires decrypt function")
	}

	// Decrypt if needed
	var decryptedContent []byte
	var err error
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
		flags: flags,
		data:  frameData,
	}, nil
}
