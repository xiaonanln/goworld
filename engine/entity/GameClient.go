package entity

import (
	"fmt"

	"github.com/xiaonanln/goworld/engine/common"
	"github.com/xiaonanln/goworld/engine/consts"
	"github.com/xiaonanln/goworld/engine/dispatchercluster"
	"github.com/xiaonanln/goworld/engine/dispatchercluster/dispatcherclient"
	"github.com/xiaonanln/goworld/engine/gwlog"
)

// GameClient represents the game Client of entity
//
// Each entity can have at most one GameClient, and GameClient can be given to other entities
type GameClient struct {
	clientid common.ClientID
	gateid   uint16
	ownerid  common.EntityID
}

// MakeGameClient creates a GameClient object using Client ID and Game ID
func MakeGameClient(clientid common.ClientID, gateid uint16) *GameClient {
	return &GameClient{
		clientid: clientid,
		gateid:   gateid,
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
	client.selectDispatcher().SendCreateEntityOnClient(client.gateid, client.clientid, entity.TypeName, entity.ID, isPlayer,
		clientData, float32(pos.X), float32(pos.Y), float32(pos.Z), float32(yaw))
}

func (client *GameClient) sendDestroyEntity(entity *Entity) {
	if client != nil {
		client.selectDispatcher().SendDestroyEntityOnClient(client.gateid, client.clientid, entity.TypeName, entity.ID)
	}
}

func (client *GameClient) call(entityID common.EntityID, method string, args []interface{}) {
	if client != nil {
		client.selectDispatcher().SendCallEntityMethodOnClient(client.gateid, client.clientid, entityID, method, args)
	}
}

// sendNotifyMapAttrChange updates MapAttr change to Client entity
func (client *GameClient) sendNotifyMapAttrChange(entityID common.EntityID, path []interface{}, key string, val interface{}) {
	if client != nil {
		client.selectDispatcher().SendNotifyMapAttrChangeOnClient(client.gateid, client.clientid, entityID, path, key, val)
	}
}

// sendNotifyMapAttrDel updates MapAttr delete to Client entity
func (client *GameClient) sendNotifyMapAttrDel(entityID common.EntityID, path []interface{}, key string) {
	if client != nil {
		client.selectDispatcher().SendNotifyMapAttrDelOnClient(client.gateid, client.clientid, entityID, path, key)
	}
}

func (client *GameClient) sendNotifyMapAttrClear(entityID common.EntityID, path []interface{}) {
	if client != nil {
		client.selectDispatcher().SendNotifyMapAttrClearOnClient(client.gateid, client.clientid, entityID, path)
	}
}

// sendNotifyListAttrChange notifies Client of ListAttr item changing
func (client *GameClient) sendNotifyListAttrChange(entityID common.EntityID, path []interface{}, index uint32, val interface{}) {
	if client != nil {
		client.selectDispatcher().SendNotifyListAttrChangeOnClient(client.gateid, client.clientid, entityID, path, index, val)
	}
}

// sendNotifyListAttrPop notify Client of ListAttr popping
func (client *GameClient) sendNotifyListAttrPop(entityID common.EntityID, path []interface{}) {
	if client != nil {
		client.selectDispatcher().SendNotifyListAttrPopOnClient(client.gateid, client.clientid, entityID, path)
	}
}

// sendNotifyListAttrAppend notify entity of ListAttr appending
func (client *GameClient) sendNotifyListAttrAppend(entityID common.EntityID, path []interface{}, val interface{}) {
	if client != nil {
		client.selectDispatcher().SendNotifyListAttrAppendOnClient(client.gateid, client.clientid, entityID, path, val)
	}
}

func (client *GameClient) sendSetClientFilterProp(key, val string) {
	if client != nil {
		client.selectDispatcher().SendSetClientFilterProp(client.gateid, client.clientid, key, val)
	}
}

func (client *GameClient) selectDispatcher() *dispatcherclient.DispatcherClient {
	if consts.DEBUG_MODE {
		if client.ownerid == "" {
			gwlog.Panicf("%s select dispatcher failed: ownerid is nil", client)
		}
	}
	return dispatchercluster.SelectByEntityID(client.ownerid)
}
