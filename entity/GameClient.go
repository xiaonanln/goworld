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
	gateid   uint16
}

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

func (client *GameClient) SendCreateEntity(entity *Entity, isPlayer bool) {
	if client == nil {
		return
	}

	var clientData map[string]interface{}
	if !isPlayer {
		clientData = entity.getAllClientData()
	} else {
		clientData = entity.getClientData()
	}

	pos := entity.aoi.pos
	yaw := entity.yaw
	dispatcher_client.GetDispatcherClientForSend().SendCreateEntityOnClient(client.gateid, client.clientid, entity.TypeName, entity.ID, isPlayer,
		clientData, float32(pos.X), float32(pos.Y), float32(pos.Z), float32(yaw))
}

func (client *GameClient) SendDestroyEntity(entity *Entity) {
	if client == nil {
		return
	}
	dispatcher_client.GetDispatcherClientForSend().SendDestroyEntityOnClient(client.gateid, client.clientid, entity.TypeName, entity.ID)
}

func (client *GameClient) call(entityID common.EntityID, method string, args ...interface{}) {
	if client == nil {
		return
	}
	dispatcher_client.GetDispatcherClientForSend().SendCallEntityMethodOnClient(client.gateid, client.clientid, entityID, method, args)
}

func (client *GameClient) SendNotifyMapAttrChange(entityID common.EntityID, path []interface{}, key string, val interface{}) {
	if client == nil {
		return
	}
	if consts.DEBUG_CLIENTS {
		gwlog.Debug("%s.SendNotifyMapAttrChange: entityID=%s, path=%s, %s=%v", client, entityID, path, key, val)
	}
	dispatcher_client.GetDispatcherClientForSend().SendNotifyMapAttrChangeOnClient(client.gateid, client.clientid, entityID, path, key, val)
}

func (client *GameClient) SendNotifyMapAttrDel(entityID common.EntityID, path []interface{}, key string) {
	if client == nil {
		return
	}
	if consts.DEBUG_CLIENTS {
		gwlog.Debug("%s.SendNotifyMapAttrDel: entityID=%s, path=%s, %s", client, entityID, path, key)
	}
	dispatcher_client.GetDispatcherClientForSend().SendNotifyMapAttrDelOnClient(client.gateid, client.clientid, entityID, path, key)
}

func (client *GameClient) SendNotifyListAttrChange(entityID common.EntityID, path []interface{}, index uint32, val interface{}) {
	if client == nil {
		return
	}
	if consts.DEBUG_CLIENTS {
		gwlog.Debug("%s.SendNotifyListAttrChange: entityID=%s, path=%s, %d=%v", client, entityID, path, index, val)
	}
	dispatcher_client.GetDispatcherClientForSend().SendNotifyListAttrChangeOnClient(client.gateid, client.clientid, entityID, path, index, val)
}

func (client *GameClient) SendNotifyListAttrPop(entityID common.EntityID, path []interface{}) {
	if client == nil {
		return
	}
	if consts.DEBUG_CLIENTS {
		gwlog.Debug("%s.SendNotifyListAttrPop: entityID=%s, path=%s", client, entityID, path)
	}
	dispatcher_client.GetDispatcherClientForSend().SendNotifyListAttrPopOnClient(client.gateid, client.clientid, entityID, path)
}

func (client *GameClient) SendNotifyListAttrAppend(entityID common.EntityID, path []interface{}, val interface{}) {
	if client == nil {
		return
	}
	if consts.DEBUG_CLIENTS {
		gwlog.Debug("%s.SendNotifyListAttrAppend: entityID=%s, path=%s, %v", client, entityID, val)
	}
	dispatcher_client.GetDispatcherClientForSend().SendNotifyListAttrAppendOnClient(client.gateid, client.clientid, entityID, path, val)
}

func (client *GameClient) UpdatePositionOnClient(entityID common.EntityID, position Position) {
	if client == nil {
		return
	}

	dispatcher_client.GetDispatcherClientForSend().SendUpdatePositionOnClient(client.gateid, client.clientid, entityID,
		float32(position.X), float32(position.Y), float32(position.Z))
}

func (client *GameClient) UpdateYawOnClient(entityID common.EntityID, yaw Yaw) {
	if client == nil {
		return
	}

	dispatcher_client.GetDispatcherClientForSend().SendUpdateYawOnClient(client.gateid, client.clientid, entityID, float32(yaw))
}
