package entity

import (
	"fmt"

	"github.com/xiaonanln/goworld/common"
	"github.com/xiaonanln/goworld/components/dispatcher/dispatcher_client"
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

func (client *GameClient) SendCreateEntity(entity *Entity) {
	dispatcher_client.GetDispatcherClientForSend().SendCreateEntityOnClient(client.serverid, client.clientid, entity.TypeName, entity.ID)
}

func (client *GameClient) SendDestroyEntity(entity *Entity) {
	dispatcher_client.GetDispatcherClientForSend().SendDestroyEntityOnClient(client.serverid, client.clientid, entity.TypeName, entity.ID)
}

func (client *GameClient) Call(method string, args ...interface{}) {

}
