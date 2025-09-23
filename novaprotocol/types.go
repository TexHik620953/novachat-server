package novaprotocol

import "hash/fnv"

const (
	saltSize = 8
	crcSize  = 4
)

// CryptFunc represents encryption/decryption function signature
type CryptFunc func([]byte) ([]byte, error)

func calculateCRC(header, content []byte) (uint32, error) {
	h := fnv.New32a()

	if _, err := h.Write(header); err != nil {
		return 0, err
	}
	if _, err := h.Write(content); err != nil {
		return 0, err
	}

	return h.Sum32(), nil
}
