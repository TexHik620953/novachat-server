package serverapi

import "github.com/google/uuid"

type Client struct {
	ID       uuid.UUID `json:"id"`
	Nickname string    `json:"nickname"`
}
type ListClientsResponse struct {
	Clients []Client `json:"clients"`
}
