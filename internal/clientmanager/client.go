package clientmanager

import (
	"fmt"
	"io"
	"novachat-server/novaprotocol"

	"github.com/google/uuid"
)

type Client interface {
	io.ReadWriteCloser
	SetEncryptionKey(key []byte)

	Encrypt(data []byte) ([]byte, error)
	Decrypt(data []byte) ([]byte, error)

	GetID() uuid.UUID

	SetInfo(nickname string)
	GetNickname() string
}

type client struct {
	id      uuid.UUID
	conn    io.ReadWriteCloser
	manager *clientManagerImpl

	encrypt novaprotocol.CryptFunc
	decrypt novaprotocol.CryptFunc

	nickname string
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
	c.encrypt, c.decrypt = novaprotocol.NewCryptoFuncs(key)
}

func (c *client) Encrypt(data []byte) ([]byte, error) {
	if c.encrypt != nil {
		return c.encrypt(data)
	}
	return nil, fmt.Errorf("no encrypt func")
}
func (c *client) Decrypt(data []byte) ([]byte, error) {
	if c.decrypt != nil {
		return c.decrypt(data)
	}
	return nil, fmt.Errorf("no decrypt func")
}

func (c *client) SetInfo(nickname string) {
	c.nickname = nickname
}
func (c *client) GetNickname() string {
	return c.nickname
}
