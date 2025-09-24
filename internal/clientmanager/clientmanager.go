package clientmanager

import (
	"io"
	"novachat-server/common/safemap"

	"github.com/google/uuid"
)

type ClientManager interface {
	NewClient(rw io.ReadWriteCloser) (Client, error)
}
type clientManagerImpl struct {
	clients safemap.Safemap[uuid.UUID, Client]
}

func NewClientManager() ClientManager {
	return &clientManagerImpl{
		clients: safemap.New[uuid.UUID, Client](),
	}
}

func (cm *clientManagerImpl) NewClient(rw io.ReadWriteCloser) (Client, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}
	c := &client{
		id:      id,
		conn:    rw,
		manager: cm,
	}
	cm.clients.Set(id, c)
	return c, nil
}
