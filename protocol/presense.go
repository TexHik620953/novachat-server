package protocol

import (
	"encoding/json"

	"github.com/google/uuid"
)

// Client to server
type PresenseRequest struct {
	Name string
}

// Server to sender client
type PresenseResponse struct {
	UserID uuid.UUID
}

// Server to everyone
type PresenseInfo struct {
	UserID uuid.UUID
	Name   string
}

func ParsePresenseRequest(msg []byte) (*PresenseRequest, error) {
	var r messageImpl[*PresenseRequest]
	if err := json.Unmarshal(msg, &r); err != nil {
		return nil, err
	}
	return r.Data, nil
}
func NewPresenseRequest(msg *PresenseRequest) ([]byte, error) {
	return NewMessage("pr_req", msg)
}

func ParsePresenseResponse(msg []byte) (*PresenseResponse, error) {
	var r messageImpl[*PresenseResponse]
	if err := json.Unmarshal(msg, &r); err != nil {
		return nil, err
	}
	return r.Data, nil
}
func NewPresenseResponse(msg *PresenseResponse) ([]byte, error) {
	return NewMessage("pr_res", msg)
}

func ParsePresenseInfo(msg []byte) (*PresenseInfo, error) {
	var r messageImpl[*PresenseInfo]
	if err := json.Unmarshal(msg, &r); err != nil {
		return nil, err
	}
	return r.Data, nil
}
func NewPresenseInfo(msg *PresenseInfo) ([]byte, error) {
	return NewMessage("pr_inf", msg)
}
