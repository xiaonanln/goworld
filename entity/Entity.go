package entity

import (
	"fmt"

	"time"

	timer "github.com/xiaonanln/goTimer"
	"github.com/xiaonanln/goworld/gwlog"
	"github.com/xiaonanln/goworld/uuid"
)

const ENTITYID_LENGTH = uuid.UUID_LENGTH

type EntityID string

func GenEntityID() EntityID {
	return EntityID(uuid.GenUUID())
}

type Entity struct {
	ID       EntityID
	TypeName string
	I        IEntity
	space    *Space

	timers map[*timer.Timer]struct{}
}

type IEntity interface {
	OnCreated()
	OnDestroy()
}

func (e *Entity) String() string {
	return fmt.Sprintf("%s<%s>", e.TypeName, e.ID)
}

func (e *Entity) Destroy() {
	gwlog.Info("%s.Destroy.", e)
	if e.space != nil {
		e.space.leave(e)
	}
	e.clearTimers()
	e.I.OnDestroy()
	entityManager.del(e.ID)
}

// Timer & Callback Management
func (e *Entity) AddCallback(d time.Duration, cb timer.CallbackFunc) {
	var t *timer.Timer
	t = timer.AddCallback(d, func() {
		delete(e.timers, t)
		cb()
	})
	e.timers[t] = struct{}{}
}

func (e *Entity) AddTimer(d time.Duration, cb timer.CallbackFunc) {
	t := timer.AddTimer(d, cb)
	e.timers[t] = struct{}{}
}

func (e *Entity) clearTimers() {
	for t := range e.timers {
		t.Cancel()
	}
	e.timers = map[*timer.Timer]struct{}{}
}

// Default Handlers
func (e *Entity) OnCreated() {
	gwlog.Debug("%s.OnCreated", e)
}

func (e *Entity) OnDestroy() {
}
