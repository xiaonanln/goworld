package msgbox

import (
	"github.com/xiaonanln/goworld/engine/common"
	"github.com/xiaonanln/goworld/engine/entity"
)

// msgbox is a reliable message box for entities
// Entity can use Msgbox.Send to send message to other entity and use Msgbox.Recv to receive messages
// Message can always be received by target entity even when target entity is not online, in which case target entity
// will receive this message when it is loaded and calls Msgbox.Receive.

// Msgbox is used to send messages among entities: e.x. Msgbox{&a.Entity}.Send(targetID, msg, callback)
type Msgbox struct {
	entity *entity.Entity
}

func (mb Msgbox) Send(targetID common.EntityID, msg interface{}, callback func(err error)) {

}

func (mb Msgbox) Recv(targetID common.EntityID, callback func(msgs []interface{}, err error)) {

}
