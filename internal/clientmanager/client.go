package clientmanager

import (
	"io"

	"github.com/google/uuid"
)

type Client interface {
	io.ReadWriteCloser
	SetEncryptionKey(key []byte)
	GetEncryptionKey() []byte
	GetID() uuid.UUID
}

type client struct {
	id      uuid.UUID
	conn    io.ReadWriteCloser
	manager *clientManagerImpl
	key     []byte
}

func (c *client) Read(p []byte) (n int, err error) {
	return c.conn.Read(p)
}
func (c *client) Write(p []byte) (n int, err error) {
	return c.conn.Write(p)
}
func (c *client) Close() error {
	c.manager.clients.Remove(c.id)
	return c.conn.Close()
}

func (c *client) GetID() uuid.UUID {
	return c.id
}

func (c *client) SetEncryptionKey(key []byte) {
	c.key = key
}
func (c *client) GetEncryptionKey() []byte {
	return c.key
}
