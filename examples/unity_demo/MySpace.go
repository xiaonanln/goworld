package main

import (
	"strconv"
	"time"

	"github.com/xiaonanln/goTimer"
	"github.com/xiaonanln/goworld"
	"github.com/xiaonanln/goworld/engine/entity"
	"github.com/xiaonanln/goworld/engine/gwlog"
)

const (
	_SPACE_DESTROY_CHECK_INTERVAL = time.Minute * 5
)

// MySpace is the custom space type
type MySpace struct {
	goworld.Space // Space type should always inherit from entity.Space

	destroyCheckTimer entity.EntityTimerID
}

// OnSpaceCreated is called when the space is created
func (space *MySpace) OnSpaceCreated() {
	// notify the SpaceService that it's ok
	space.EnableAOI(100)

	goworld.CallServiceShardKey("SpaceService", strconv.Itoa(space.Kind), "NotifySpaceLoaded", space.Kind, space.ID)
	space.AddTimer(time.Second*5, "DumpEntityStatus")
	space.AddTimer(time.Second*5, "SummonMonsters")
	//M := 10
	//for i := 0; i < M; i++ {
	//	space.CreateEntity("Monster", entity.Vector3{})
	//}
}

func (space *MySpace) DumpEntityStatus() {
	space.ForEachEntity(func(e *entity.Entity) {
		gwlog.Debugf(">>> %s @ position %s, neighbors=%d", e, e.GetPosition(), len(e.InterestedIn))
	})
}

func (space *MySpace) SummonMonsters() {
	if space.CountEntities("Monster") < space.CountEntities("Player")*2 {
		space.CreateEntity("Monster", entity.Vector3{})
	}
}

// OnEntityEnterSpace is called when entity enters space
func (space *MySpace) OnEntityEnterSpace(entity *entity.Entity) {
	if entity.TypeName == "Player" {
		space.onPlayerEnterSpace(entity)
	}
}

func (space *MySpace) onPlayerEnterSpace(entity *entity.Entity) {
	gwlog.Debugf("Player %s enter space %s, total avatar count %d", entity, space, space.CountEntities("Player"))
	space.clearDestroyCheckTimer()
}

// OnEntityLeaveSpace is called when entity leaves space
func (space *MySpace) OnEntityLeaveSpace(entity *entity.Entity) {
	if entity.TypeName == "Player" {
		space.onPlayerLeaveSpace(entity)
	}
}

func (space *MySpace) onPlayerLeaveSpace(entity *entity.Entity) {
	gwlog.Infof("Player %s leave space %s, left avatar count %d", entity, space, space.CountEntities("Player"))
	if space.CountEntities("Player") == 0 {
		// no avatar left, start destroying space
		space.setDestroyCheckTimer()
	}
}

func (space *MySpace) setDestroyCheckTimer() {
	if space.destroyCheckTimer != 0 {
		return
	}

	space.destroyCheckTimer = space.AddTimer(_SPACE_DESTROY_CHECK_INTERVAL, "CheckForDestroy")
}

// CheckForDestroy checks if the space should be destroyed
func (space *MySpace) CheckForDestroy() {
	avatarCount := space.CountEntities("Player")
	if avatarCount != 0 {
		gwlog.Panicf("Player count should be 0, but is %d", avatarCount)
	}

	goworld.CallServiceShardKey("SpaceService", strconv.Itoa(space.Kind), "RequestDestroy", space.Kind, space.ID)
}

func (space *MySpace) clearDestroyCheckTimer() {
	if space.destroyCheckTimer == 0 {
		return
	}

	space.CancelTimer(space.destroyCheckTimer)
	space.destroyCheckTimer = 0
}

// ConfirmRequestDestroy is called by SpaceService to confirm that the space
func (space *MySpace) ConfirmRequestDestroy(ok bool) {
	if ok {
		if space.CountEntities("Player") != 0 {
			gwlog.Panicf("%s ConfirmRequestDestroy: avatar count is %d", space, space.CountEntities("Player"))
		}
		space.Destroy()
	}
}

// OnGameReady is called when the game server is ready
func (space *MySpace) OnGameReady() {
	timer.AddCallback(time.Millisecond*1000, checkServerStarted)
}

func checkServerStarted() {
	ok := isAllServicesReady()
	gwlog.Infof("checkServerStarted: %v", ok)
	if ok {
		onAllServicesReady()
	} else {
		timer.AddCallback(time.Millisecond*1000, checkServerStarted)
	}
}

func isAllServicesReady() bool {
	for _, serviceName := range _SERVICE_NAMES {
		if !goworld.CheckServiceEntitiesReady(serviceName) {
			gwlog.Infof("%s entities are not ready ...", serviceName)
			return false
		}
	}
	return true
}

func onAllServicesReady() {
	gwlog.Infof("All services are ready!")
}
