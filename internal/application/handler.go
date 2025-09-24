package application

import (
	"fmt"
	"io"
	"log"
	"novachat-server/internal/clientmanager"
	"novachat-server/novaprotocol"

	"github.com/google/uuid"
)

// Constants for better maintainability
const (
	maxKeyExchangeRetryAttempts = 3
)

func respondJson(client clientmanager.Client, jsonData []byte) error {
	l1, err := novaprotocol.NewL1Frame(novaprotocol.L1FlagIsJson, jsonData).Build(nil)
	if err != nil {
		return err
	}
	return novaprotocol.NewL0Frame(novaprotocol.L0FlagIsEncrypted, client.GetID(), l1).Write(client, client.Encrypt)
}

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

	err = sendWelcomeInviteMessage(client)
	if err != nil {
		return fmt.Errorf("welcome failed: %w", err)
	}

	cInfo, err := recvWelcomeAcceptMessage(client)
	if err != nil {
		return fmt.Errorf("welcome accept failed: %w", err)
	}
	client.SetInfo(cInfo.Nickname)

	// Send welcome message
	log.Printf("successfully established secure connection with client: %s", client.GetID().String())

	// Main messaging cycle
	for {
		l0frame, err := novaprotocol.ReadL0Frame(client, client.Decrypt)
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
			app.routeAPI(client, l0frame)

		} else {
			// Unicast
			target, ex := app.clientManager.GetClient(l0frame.GetDestination())
			if !ex {
				log.Printf("unicast target not found")
				continue
			}

			err = l0frame.Write(target, target.Encrypt)
			if err != nil {
				log.Printf("failed to unicast message: %v", err)
				continue
			}
		}
	}
	return nil
}
