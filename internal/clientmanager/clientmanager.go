package clientmanager

import (
	"io"
	"novachat-server/common/safemap"

	"github.com/google/uuid"
)

type ClientManager interface {
	NewClient(rw io.ReadWriteCloser) (Client, error)
	GetClient(id uuid.UUID) (Client, bool)
	ListClients() []Client
}
type clientManagerImpl struct {
	clients safemap.Safemap[uuid.UUID, Client]
}

func NewClientManager() ClientManager {
	return &clientManagerImpl{
		clients: safemap.New[uuid.UUID, Client](),
	}
}

func (cm *clientManagerImpl) GetClient(id uuid.UUID) (Client, bool) {
	return cm.clients.Get(id)
}
func (cm *clientManagerImpl) ListClients() []Client {
	clients := make([]Client, 0)
	cm.clients.Foreach(func(u uuid.UUID, c Client) {
		clients = append(clients, c)
	})
	return clients
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
