package game

import (
	"github.com/xiaonanln/goworld/components/dispatcher/dispatcher_client"
	"github.com/xiaonanln/goworld/entity"
)

func CreateEntity(typeName string) {
	entityID := entity.CreateEntity(typeName)
	// tell the dispatcher about the entity creation
	dispatcher_client.GetDispatcherClientForSend().NotifyCreateEntity(entityID)
}
