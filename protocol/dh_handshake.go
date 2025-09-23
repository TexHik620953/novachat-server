package protocol

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"math/big"
)

var (
	prime2048, _  = new(big.Int).SetString("FFFFFFFFFFFFFFFFC90FDAA22168C234C4C6628B80DC1CD129024E088A67CC74020BBEA63B139B22514A08798E3404DDEF9519B3CD3A431B302B0A6DF25F14374FE1356D6D51C245E485B576625E7EC6F44C42E9A637ED6B0BFF5CB6F406B7EDEE386BFB5A899FA5AE9F24117C4B1FE649286651ECE45B3DC2007CB8A163BF0598DA48361C55D39A69163FA8FD24CF5F83655D23DCA3AD961C62F356208552BB9ED529077096966D670C354E4ABC9804F1746C08CA18217C32905E462E36CE3BE39E772C180E86039B2783A2EC07A28FB5C55DF06F4C52C9DE2BCBF6955817183995497CEA956AE515D2261898FA051015728E5A8AACAA68FFFFFFFFFFFFFFFF", 16)
	generator2048 = big.NewInt(2)
)

// Генерация ключевой пары
func GenerateKeyPair() (private, public *big.Int, err error) {
	private, err = rand.Int(rand.Reader, prime2048)
	if err != nil {
		return nil, nil, err
	}
	public = new(big.Int).Exp(generator2048, private, prime2048)
	return private, public, nil
}

// Вычисление общего ключа
func ComputeSharedKey(private, peerPublic *big.Int) []byte {
	sharedSecret := new(big.Int).Exp(peerPublic, private, prime2048)
	// Используем SHA256 для получения 32-байтного ключа (AES-256)
	key := sha256.Sum256(sharedSecret.Bytes())
	return key[:]
}

type PublicKeyMessage struct {
	PublicKey string `json:"public_key"`
}

func ParsePublicKeyMessage(msg []byte) (*PublicKeyMessage, error) {
	var r messageImpl[*PublicKeyMessage]
	if err := json.Unmarshal(msg, &r); err != nil {
		return nil, err
	}
	return r.Data, nil
}
func NewPublicKeyMessage(msg *PublicKeyMessage) ([]byte, error) {
	return NewMessage("pub", msg)
}
