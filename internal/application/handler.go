package application

import (
	"fmt"
	"io"
	"log"
	"novachat-server/novaprotocol"

	"github.com/google/uuid"
)

// Constants for better maintainability
const (
	maxRetryAttempts = 3
	keySizeBits      = 256 // For key derivation
)

// connectionHandler manages the entire client connection lifecycle
func (app *Application) connectionHandler(rw io.ReadWriteCloser) error {
	client, err := app.clientManager.NewClient(rw)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			log.Printf("Error closing client: %v", err)
		}
	}()

	// Perform key exchange
	encryptionKey, err := keyExchange(client)
	if err != nil {
		return fmt.Errorf("key exchange failed: %w", err)
	}
	// Set encryption key for the client
	client.SetEncryptionKey(encryptionKey)
	encrypt, decrypt := novaprotocol.NewCryptoFuncs(encryptionKey)

	err = sendWelcomeMessage(client, encrypt)
	if err != nil {
		return fmt.Errorf("welcome failed: %w", err)
	}

	// Send welcome message
	log.Printf("successfully established secure connection with client: %s", client.GetID().String())

	// Main messaging cycle
	for {
		l0frame, err := novaprotocol.ReadL0Frame(client, decrypt)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			log.Printf("failed to read l0 frame: %v", err)
		}
		if l0frame.GetOrigin() != client.GetID() {
			log.Printf("invalid packet source")
			continue
		}

		if l0frame.GetDestination() == uuid.Nil {
			// Message for server
		}
	}
	return nil
}
