package application

import (
	"fmt"
	"novachat-server/common/linq"
	"novachat-server/internal/clientmanager"
	"novachat-server/novaprotocol"
	"novachat-server/novaprotocol/serverapi"
)

func (app *Application) routeAPI(client clientmanager.Client, l0Frame *novaprotocol.NovaFrameL0) error {
	// Message for server, so l1 should be unencrypted
	l1Frame, err := novaprotocol.ParseL1Frame(l0Frame.GetData(), nil)
	if err != nil {
		return fmt.Errorf("failed to parse l1 frame: %v", err)
	}

	if l1Frame.GetFlags()&novaprotocol.L1FlagIsJson != 0 {
		// Json message
		msgType, err := novaprotocol.ParseJsonMessageType(l1Frame.GetData())
		if err != nil {
			return fmt.Errorf("failed to parse msg type: %w", err)
		}
		switch msgType {
		case novaprotocol.MSG_LIST_CONN:
			err = app.listConnections(client)
		}
		if err != nil {
			return fmt.Errorf("failed to execute api method: %w", err)
		}

	} else if l1Frame.GetFlags()&novaprotocol.L1FlagIsFile != 0 {
		// File message
	}

	return nil
}

func (app *Application) listConnections(client clientmanager.Client) error {
	resp := linq.Select(app.clientManager.ListClients(), func(c clientmanager.Client) *serverapi.Client {
		return &serverapi.Client{
			ID:       c.GetID(),
			Nickname: c.GetNickname(),
		}
	})
	msg, err := novaprotocol.NewJsonMessage(novaprotocol.MSG_LIST_CONN, resp)
	if err != nil {
		return err
	}
	return respondJson(client, msg)
}
