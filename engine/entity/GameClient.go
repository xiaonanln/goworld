package entity

import (
	"fmt"

	"github.com/xiaonanln/goworld/components/dispatcher/dispatcherclient"
	"github.com/xiaonanln/goworld/engine/common"
	"github.com/xiaonanln/goworld/engine/consts"
	"github.com/xiaonanln/goworld/engine/gwlog"
)

// GameClient represents the game client of entity
//
// Each entity can have at most one GameClient, and GameClient can be given to other entities
type GameClient struct {
	clientid common.ClientID
	gateid   uint16
}

// MakeGameClient creates a GameClient object using Client ID and Game ID
func MakeGameClient(clientid common.ClientID, gid uint16) *GameClient {
	return &GameClient{
		clientid: clientid,
		gateid:   gid,
	}
}

func (client *GameClient) String() string {
	if client == nil {
		return "GameClient<nil>"
	}
	return fmt.Sprintf("GameClient<%s@%d>", client.clientid, client.gateid)
}

func (client *GameClient) sendCreateEntity(entity *Entity, isPlayer bool) {
	if client == nil {
		return
	}

	var clientData map[string]interface{}
	if !isPlayer {
		clientData = entity.getAllClientData()
	} else {
		clientData = entity.getClientData()
	}

	pos := entity.Position
	yaw := entity.yaw
	dispatcherclient.GetDispatcherClientForSend().SendCreateEntityOnClient(client.gateid, client.clientid, entity.TypeName, entity.ID, isPlayer,
		clientData, float32(pos.X), float32(pos.Y), float32(pos.Z), float32(yaw))
}

func (client *GameClient) sendDestroyEntity(entity *Entity) {
	if client == nil {
		return
	}
	dispatcherclient.GetDispatcherClientForSend().SendDestroyEntityOnClient(client.gateid, client.clientid, entity.TypeName, entity.ID)
}

func (client *GameClient) call(entityID common.EntityID, method string, args []interface{}) {
	if client == nil {
		return
	}
	dispatcherclient.GetDispatcherClientForSend().SendCallEntityMethodOnClient(client.gateid, client.clientid, entityID, method, args)
}

// sendNotifyMapAttrChange updates MapAttr change to client entity
func (client *GameClient) sendNotifyMapAttrChange(entityID common.EntityID, path []interface{}, key string, val interface{}) {
	if client == nil {
		return
	}
	if consts.DEBUG_CLIENTS {
		gwlog.Debugf("%s.sendNotifyMapAttrChange: entityID=%s, path=%s, %s=%v", client, entityID, path, key, val)
	}
	dispatcherclient.GetDispatcherClientForSend().SendNotifyMapAttrChangeOnClient(client.gateid, client.clientid, entityID, path, key, val)
}

// sendNotifyMapAttrDel updates MapAttr delete to client entity
func (client *GameClient) sendNotifyMapAttrDel(entityID common.EntityID, path []interface{}, key string) {
	if client == nil {
		return
	}
	if consts.DEBUG_CLIENTS {
		gwlog.Debugf("%s.sendNotifyMapAttrDel: entityID=%s, path=%s, %s", client, entityID, path, key)
	}
	dispatcherclient.GetDispatcherClientForSend().SendNotifyMapAttrDelOnClient(client.gateid, client.clientid, entityID, path, key)
}

// sendNotifyListAttrChange notifies client of ListAttr item changing
func (client *GameClient) sendNotifyListAttrChange(entityID common.EntityID, path []interface{}, index uint32, val interface{}) {
	if client == nil {
		return
	}
	if consts.DEBUG_CLIENTS {
		gwlog.Debugf("%s.sendNotifyListAttrChange: entityID=%s, path=%s, %d=%v", client, entityID, path, index, val)
	}
	dispatcherclient.GetDispatcherClientForSend().SendNotifyListAttrChangeOnClient(client.gateid, client.clientid, entityID, path, index, val)
}

// sendNotifyListAttrPop notify client of ListAttr popping
func (client *GameClient) sendNotifyListAttrPop(entityID common.EntityID, path []interface{}) {
	if client == nil {
		return
	}
	if consts.DEBUG_CLIENTS {
		gwlog.Debugf("%s.sendNotifyListAttrPop: entityID=%s, path=%s", client, entityID, path)
	}
	dispatcherclient.GetDispatcherClientForSend().SendNotifyListAttrPopOnClient(client.gateid, client.clientid, entityID, path)
}

// sendNotifyListAttrAppend notify entity of ListAttr appending
func (client *GameClient) sendNotifyListAttrAppend(entityID common.EntityID, path []interface{}, val interface{}) {
	if client == nil {
		return
	}
	if consts.DEBUG_CLIENTS {
		gwlog.Debugf("%s.sendNotifyListAttrAppend: entityID=%s, path=%s, %v", client, entityID, val)
	}
	dispatcherclient.GetDispatcherClientForSend().SendNotifyListAttrAppendOnClient(client.gateid, client.clientid, entityID, path, val)
}
