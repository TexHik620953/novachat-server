package novaprotocol

import (
	"encoding/binary"
	"fmt"

	"github.com/google/uuid"
)

const (
	FPackTypeFileStart byte = iota
	FPackTypeFileBlock
	FPackTypeFileRequest
)

type FileStartFrameParams struct {
	FileSize    uint32
	BlocksCount uint16
	FileName    string
	FileID      uuid.UUID
	FileHash    [32]byte
}

type FileBlockFrameParams struct {
	BlockIdx uint16
	FileID   uuid.UUID
	Data     []byte
}

type FileRequestBlockFrameParams struct {
	BlockIdx uint16
	FileID   uuid.UUID
}

func NewFileStartFrame(params FileStartFrameParams) ([]byte, error) {
	if len(params.FileName) > 255 {
		return nil, fmt.Errorf("file name is too long")
	}

	buff := make([]byte, 0, 1+1+4+2+1+len(params.FileName)+16+32)
	buff = append(buff, FPackTypeFileStart)
	buff = append(buff, 0)
	buff = binary.LittleEndian.AppendUint32(buff, params.FileSize)
	buff = binary.LittleEndian.AppendUint16(buff, params.BlocksCount)
	buff = append(buff, byte(len(params.FileName)))
	buff = append(buff, []byte(params.FileName)...)
	buff = append(buff, params.FileID[:]...)
	buff = append(buff, params.FileHash[:]...)
	return buff, nil
}

func ParseFileStartFrame(data []byte) (FileStartFrameParams, error) {
	if len(data) < 1+1+4+2+1+16+32 {
		return FileStartFrameParams{}, fmt.Errorf("frame too short")
	}

	if data[0] != FPackTypeFileStart {
		return FileStartFrameParams{}, fmt.Errorf("invalid frame type")
	}

	// Пропускаем тип пакета и резервный байт
	offset := 2

	fileSize := binary.LittleEndian.Uint32(data[offset:])
	offset += 4

	blocksCount := binary.LittleEndian.Uint16(data[offset:])
	offset += 2

	fileNameLen := int(data[offset])
	offset += 1

	if len(data) < offset+fileNameLen+16+32 {
		return FileStartFrameParams{}, fmt.Errorf("frame too short for file name and metadata")
	}

	fileName := string(data[offset : offset+fileNameLen])
	offset += fileNameLen

	fileID, err := uuid.FromBytes(data[offset : offset+16])
	if err != nil {
		return FileStartFrameParams{}, fmt.Errorf("invalid file ID: %w", err)
	}
	offset += 16

	var fileHash [32]byte
	copy(fileHash[:], data[offset:offset+32])

	return FileStartFrameParams{
		FileSize:    fileSize,
		BlocksCount: blocksCount,
		FileName:    fileName,
		FileID:      fileID,
		FileHash:    fileHash,
	}, nil
}

func NewFileBlockFrame(blockIdx uint16, fileID uuid.UUID, data []byte) []byte {
	buff := make([]byte, 0, 1+2+16+len(data))
	buff = append(buff, FPackTypeFileBlock)
	buff = binary.LittleEndian.AppendUint16(buff, blockIdx)
	buff = append(buff, fileID[:]...)
	buff = append(buff, data...)
	return buff
}

func ParseFileBlockFrame(data []byte) (FileBlockFrameParams, error) {
	if len(data) < 1+2+16 {
		return FileBlockFrameParams{}, fmt.Errorf("frame too short")
	}

	if data[0] != FPackTypeFileBlock {
		return FileBlockFrameParams{}, fmt.Errorf("invalid frame type")
	}

	offset := 1
	blockIdx := binary.LittleEndian.Uint16(data[offset:])
	offset += 2

	fileID, err := uuid.FromBytes(data[offset : offset+16])
	if err != nil {
		return FileBlockFrameParams{}, fmt.Errorf("invalid file ID: %w", err)
	}
	offset += 16

	blockData := make([]byte, len(data)-offset)
	copy(blockData, data[offset:])

	return FileBlockFrameParams{
		BlockIdx: blockIdx,
		FileID:   fileID,
		Data:     blockData,
	}, nil
}

func NewFileRequestBlockFrame(blockIdx uint16, fileID uuid.UUID) []byte {
	buff := make([]byte, 0, 1+1+16+2)
	buff = append(buff, FPackTypeFileRequest)
	buff = append(buff, 0)
	buff = append(buff, fileID[:]...)
	buff = binary.LittleEndian.AppendUint16(buff, blockIdx)
	return buff
}

func ParseFileRequestBlockFrame(data []byte) (FileRequestBlockFrameParams, error) {
	if len(data) != 1+1+16+2 {
		return FileRequestBlockFrameParams{}, fmt.Errorf("invalid frame length")
	}

	if data[0] != FPackTypeFileRequest {
		return FileRequestBlockFrameParams{}, fmt.Errorf("invalid frame type")
	}

	// Пропускаем тип пакета и резервный байт
	offset := 2

	fileID, err := uuid.FromBytes(data[offset : offset+16])
	if err != nil {
		return FileRequestBlockFrameParams{}, fmt.Errorf("invalid file ID: %w", err)
	}
	offset += 16

	blockIdx := binary.LittleEndian.Uint16(data[offset:])

	return FileRequestBlockFrameParams{
		BlockIdx: blockIdx,
		FileID:   fileID,
	}, nil
}
