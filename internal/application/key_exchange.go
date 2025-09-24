package application

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"novachat-server/internal/clientmanager"
	"novachat-server/novaprotocol"
	"novachat-server/novaprotocol/handshake"

	"github.com/google/uuid"
)

// keyExchange performs Diffie-Hellman key exchange with mutual authentication
func keyExchange(rw clientmanager.Client) ([]byte, error) {
	for attempt := 0; attempt < maxRetryAttempts; attempt++ {
		key, err := performSingleKeyExchange(rw)
		if err != nil {
			log.Printf("Key exchange attempt %d failed: %v", attempt+1, err)
			continue
		}
		return key, nil
	}

	return nil, fmt.Errorf("key exchange failed after %d attempts", maxRetryAttempts)
}

// performSingleKeyExchange handles one iteration of the key exchange protocol
func performSingleKeyExchange(rw clientmanager.Client) ([]byte, error) {
	// Generate key pair
	privateKey, publicKey, err := handshake.GenerateKeyPair(handshake.Generator2048, handshake.Prime2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key pair: %w", err)
	}

	// Generate random challenge
	challenge := rand.Text()

	// Send our public key to client
	if err := sendPublicKeyMessage(rw, publicKey, challenge); err != nil {
		return nil, fmt.Errorf("failed to send public key: %w", err)
	}

	// Receive and validate client's public key
	clientPublicKey, clientKeyHash, err := receivePublicKeyMessage(rw)
	if err != nil {
		return nil, fmt.Errorf("failed to receive public key: %w", err)
	}
	// Compute shared secret
	sharedSecret := handshake.ComputeSharedKey(privateKey, clientPublicKey)
	// Validate challenge
	if handshake.ComputeKeyHash(challenge, sharedSecret) != clientKeyHash {
		return nil, fmt.Errorf("key hash missmatch")
	}
	return sharedSecret, nil
}

// sendPublicKeyMessage sends our public key information to the client
func sendPublicKeyMessage(rw clientmanager.Client, publicKey *big.Int, challenge string) error {
	messageData, err := novaprotocol.NewJsonMessage(novaprotocol.MSG_DH_PUB, &handshake.PublicKeyServer2Client{
		G:         handshake.Generator2048.Text(62),
		P:         handshake.Prime2048.Text(62),
		Pub:       publicKey.Text(62),
		Challenge: challenge,
	})
	if err != nil {
		return fmt.Errorf("failed to create public key message: %w", err)
	}

	l1frameData, err := novaprotocol.NewL1Frame(novaprotocol.L1FlagIsJson, messageData).Build(nil)
	if err != nil {
		return fmt.Errorf("failed to create l1 frame: %w", err)
	}
	frame := novaprotocol.NewL0Frame(novaprotocol.L0FlagNone, uuid.Nil, l1frameData)
	if err := frame.Write(rw, nil); err != nil {
		return fmt.Errorf("failed to write l0 frame: %w", err)
	}

	return nil
}

// receivePublicKeyMessage receives and validates the client's public key
func receivePublicKeyMessage(rw clientmanager.Client) (*big.Int, string, error) {
	l0frame, err := novaprotocol.ReadL0Frame(rw, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read l0 frame: %w", err)
	}

	l1frame, err := novaprotocol.ParseL1Frame(l0frame.GetData(), nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read l1 frame: %w", err)
	}
	if l1frame.GetFlags()&novaprotocol.L1FlagIsJson == 0 {
		return nil, "", fmt.Errorf("invalid data type")
	}

	// Validate message type
	messageType, err := novaprotocol.ParseJsonMessageType(l1frame.GetData())
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse message type: %w", err)
	}

	if messageType != novaprotocol.MSG_DH_PUB {
		return nil, "", fmt.Errorf("unexpected message type: expected %s, got %s",
			novaprotocol.MSG_DH_PUB, messageType)
	}

	// Parse message content
	publicKeyMsg, err := novaprotocol.ParseJsonMessage[handshake.PublicKeyClient2Server](l1frame.GetData())
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse public key message: %w", err)
	}

	// Convert string representations back to big.Int
	clientPublicKey, ok := new(big.Int).SetString(publicKeyMsg.Pub, 62)
	if !ok || clientPublicKey == nil {
		return nil, "", fmt.Errorf("invalid client public key format")
	}

	return clientPublicKey, publicKeyMsg.Hash, nil
}

func sendWelcomeMessage(client clientmanager.Client, encrypt novaprotocol.CryptFunc) error {
	messageData, err := novaprotocol.NewJsonMessage(novaprotocol.MSG_WELCOME, &handshake.WelcomeServer2Client{
		UserID: client.GetID(),
	})
	if err != nil {
		return fmt.Errorf("failed to create public key message: %w", err)
	}

	l1frame, err := novaprotocol.NewL1Frame(novaprotocol.L1FlagIsJson|novaprotocol.L1FlagIsEncrypted, messageData).Build(encrypt)
	if err != nil {
		return fmt.Errorf("failed to create l1 frame: %w", err)
	}

	frame := novaprotocol.NewL0Frame(novaprotocol.L0FlagIsEncrypted, uuid.Nil, l1frame)
	if err := frame.Write(client, nil); err != nil {
		return fmt.Errorf("failed to write l0 frame: %w", err)
	}

	return nil
}
