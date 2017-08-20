package main

import (
	"time"

	"github.com/xiaonanln/goworld/engine/entity"
)

// Monster type
type Monster struct {
	entity.Entity   // Entity type should always inherit entity.Entity
	movingToTarget  *entity.Entity
	attackingTarget *entity.Entity
	lastTickTime    time.Time
}

func (monster *Monster) OnEnterSpace() {
	monster.AddTimer(time.Millisecond*100, "AI")
	monster.lastTickTime = time.Now()
	monster.AddTimer(time.Millisecond*30, "Tick")
}

func (monster *Monster) AI() {
	var nearestPlayer *entity.Entity
	for entity := range monster.Neighbors() {

		if entity.TypeName != "Player" {
			continue
		}

		if nearestPlayer == nil || nearestPlayer.DistanceTo(&monster.Entity) > entity.DistanceTo(&monster.Entity) {
			nearestPlayer = entity
		}
	}

	if nearestPlayer == nil {
		monster.Idling()
		return
	}

	if nearestPlayer.DistanceTo(&monster.Entity) > monster.GetAttackRange() {
		monster.MovingTo(nearestPlayer)
	} else {
		monster.Attacking(nearestPlayer)
	}
}

func (monster *Monster) Tick() {
	if monster.attackingTarget != nil && monster.IsNeighbor(monster.attackingTarget) {
		monster.FaceTo(monster.attackingTarget)
		return
	}

	if monster.movingToTarget != nil && monster.IsNeighbor(monster.movingToTarget) {
		mypos := monster.GetPosition()
		direction := monster.movingToTarget.GetPosition().Sub(mypos)
		direction.Y = 0

		t := direction.Normalized().Mul(monster.GetSpeed() * 30 / 1000.0)
		monster.SetPosition(mypos.Add(t))
		monster.FaceTo(monster.movingToTarget)
		return
	}
}

func (monster *Monster) GetSpeed() entity.Coord {
	return 2
}

func (monster *Monster) GetAttackRange() entity.Coord {
	return 5
}

func (monster *Monster) Idling() {
	monster.movingToTarget = nil
	monster.attackingTarget = nil
}

func (monster *Monster) MovingTo(player *entity.Entity) {
	monster.movingToTarget = player
	monster.attackingTarget = nil
}

func (monster *Monster) Attacking(player *entity.Entity) {
	monster.movingToTarget = nil
	monster.attackingTarget = player
}
