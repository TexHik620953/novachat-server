package handshake

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"math/big"

	"github.com/google/uuid"
)

var (
	Prime2048, _  = new(big.Int).SetString("FFFFFFFFFFFFFFFFC90FDAA22168C234C4C6628B80DC1CD129024E088A67CC74020BBEA63B139B22514A08798E3404DDEF9519B3CD3A431B302B0A6DF25F14374FE1356D6D51C245E485B576625E7EC6F44C42E9A637ED6B0BFF5CB6F406B7EDEE386BFB5A899FA5AE9F24117C4B1FE649286651ECE45B3DC2007CB8A163BF0598DA48361C55D39A69163FA8FD24CF5F83655D23DCA3AD961C62F356208552BB9ED529077096966D670C354E4ABC9804F1746C08CA18217C32905E462E36CE3BE39E772C180E86039B2783A2EC07A28FB5C55DF06F4C52C9DE2BCBF6955817183995497CEA956AE515D2261898FA051015728E5A8AACAA68FFFFFFFFFFFFFFFF", 16)
	Generator2048 = big.NewInt(2)
)

// Генерация ключевой пары
func GenerateKeyPair(g, p *big.Int) (private, public *big.Int, err error) {
	private, err = rand.Int(rand.Reader, Prime2048)
	if err != nil {
		return nil, nil, err
	}
	public = new(big.Int).Exp(g, p, Prime2048)
	return private, public, nil
}

// Вычисление общего ключа
func ComputeSharedKey(private, peerPublic *big.Int) []byte {
	sharedSecret := new(big.Int).Exp(peerPublic, private, Prime2048)
	// Используем SHA256 для получения 32-байтного ключа (AES-256)
	key := sha256.Sum256(sharedSecret.Bytes())
	return key[:]
}

func ComputeKeyHash(challenge string, key []byte) string {
	h := sha256.New()
	h.Write([]byte(challenge))
	h.Write(key)
	return hex.EncodeToString(h.Sum(nil))
}

type PublicKeyServer2Client struct {
	G         string `json:"g,omitempty"`
	P         string `json:"p,omitempty"`
	Pub       string `json:"pub,omitempty"`
	Challenge string `json:"clng,omitempty"`
}

type PublicKeyClient2Server struct {
	Pub  string `json:"pub,omitempty"`
	Hash string `json:"hash,omitempty"`
}

type WelcomeServer2Client struct {
	UserID uuid.UUID
}
