package novaprotocol

import "hash/fnv"

const (
	saltSize = 8
	crcSize  = 4
)

const (
	// Client->Server|Server->Client
	MSG_DH_PUB = "dh_pub" // The only unencrypted message

	MSG_WELCOME_INVITE = "srv_welcome_invite"
	MSG_WELCOME_ACCEPT = "srv_welcome_accept"

	MSG_NEW_CONNECTION  = "srv_new_conn"
	MSG_CONNECTION_LOST = "src_conn_lost"
	MSG_LIST_CONN       = "srv_conn_list"
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
