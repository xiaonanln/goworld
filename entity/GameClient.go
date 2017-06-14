package entity

import (
	"fmt"

	"github.com/xiaonanln/goworld/common"
)

type GameClient struct {
	clientid common.ClientID
	serverid uint16
}

func MakeGameClient(clientid common.ClientID, sid uint16) *GameClient {
	return &GameClient{
		clientid: clientid,
		serverid: sid,
	}
}

func (client *GameClient) String() string {
	return fmt.Sprintf("GameClient<%s@%d>", client.clientid, client.serverid)
}
