package entity

import (
	"fmt"

	"github.com/xiaonanln/goworld/common"
	"github.com/xiaonanln/goworld/components/dispatcher/dispatcher_client"
	"github.com/xiaonanln/goworld/consts"
	"github.com/xiaonanln/goworld/gwlog"
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
	if client == nil {
		return "GameClient<nil>"
	}
	return fmt.Sprintf("GameClient<%s@%d>", client.clientid, client.serverid)
}

func (client *GameClient) SendCreateEntity(entity *Entity, isPlayer bool) {
	if client == nil {
		return
	}

	var clientData map[string]interface{}
	if !isPlayer {
		clientData = entity.getClientData()
	} else {
		clientData = entity.getAllClientData()
	}

	dispatcher_client.GetDispatcherClientForSend().SendCreateEntityOnClient(client.serverid, client.clientid, entity.TypeName, entity.ID, isPlayer, clientData)
}

func (client *GameClient) SendDestroyEntity(entity *Entity) {
	if client == nil {
		return
	}
	dispatcher_client.GetDispatcherClientForSend().SendDestroyEntityOnClient(client.serverid, client.clientid, entity.TypeName, entity.ID)
}

func (client *GameClient) call(entityID common.EntityID, method string, args ...interface{}) {
	if client == nil {
		return
	}
	dispatcher_client.GetDispatcherClientForSend().SendCallEntityMethodOnClient(client.serverid, client.clientid, entityID, method, args)
}

func (client *GameClient) SendNotifyAttrChange(entityID common.EntityID, path []string, key string, val interface{}) {
	if client == nil {
		return
	}
	if consts.DEBUG_CLIENTS {
		gwlog.Debug("%s.SendNotifyAttrChange: entityID=%s, path=%s, %s=%v", client, entityID, path, key, val)
	}
	dispatcher_client.GetDispatcherClientForSend().SendNotifyAttrChangeOnClient(client.serverid, client.clientid, entityID, path, key, val)
}

func (client *GameClient) SendNotifyAttrDel(entityID common.EntityID, path []string, key string) {
	if client == nil {
		return
	}
	if consts.DEBUG_CLIENTS {
		gwlog.Debug("%s.SendNotifyAttrDel: entityID=%s, path=%s, %s", client, entityID, path, key)
	}
	dispatcher_client.GetDispatcherClientForSend().SendNotifyAttrDelOnClient(client.serverid, client.clientid, entityID, path, key)
}
