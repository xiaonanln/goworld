package main

import (
	"time"

	"github.com/xiaonanln/goworld/engine/entity"
	"github.com/xiaonanln/goworld/engine/gwlog"
)

// Monster type
type Monster struct {
	entity.Entity   // Entity type should always inherit entity.Entity
	movingToTarget  *entity.Entity
	attackingTarget *entity.Entity
	lastTickTime    time.Time

	attackCD       time.Duration
	lastAttackTime time.Time
}

func (monster *Monster) DescribeEntityType(desc *entity.EntityTypeDesc) {
	desc.SetUseAOI(true)
	desc.DefineAttr("name", "AllClients")
	desc.DefineAttr("lv", "AllClients")
	desc.DefineAttr("hp", "AllClients")
	desc.DefineAttr("hpmax", "AllClients")
	desc.DefineAttr("action", "AllClients")
}

func (monster *Monster) OnEnterSpace() {
	monster.setDefaultAttrs()
	monster.AddTimer(time.Millisecond*100, "AI")
	monster.lastTickTime = time.Now()
	monster.AddTimer(time.Millisecond*30, "Tick")
}

func (monster *Monster) setDefaultAttrs() {
	monster.Attrs.SetDefaultStr("name", "minion")
	monster.Attrs.SetDefaultInt("lv", 1)
	monster.Attrs.SetDefaultInt("hpmax", 100)
	monster.Attrs.SetDefaultInt("hp", 100)
	monster.Attrs.SetDefaultStr("action", "idle")

	monster.attackCD = time.Second
	monster.lastAttackTime = time.Now()
}

func (monster *Monster) AI() {
	var nearestPlayer *entity.Entity
	for entity := range monster.Neighbors {

		if entity.TypeName != "Player" {
			continue
		}

		if entity.GetInt("hp") <= 0 {
			// dead
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
		now := time.Now()
		if !now.Before(monster.lastAttackTime.Add(monster.attackCD)) {
			monster.FaceTo(monster.attackingTarget)
			monster.attack(monster.attackingTarget.I.(*Player))
			monster.lastAttackTime = now
		}
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
	return 3
}

func (monster *Monster) Idling() {
	if monster.movingToTarget == nil && monster.attackingTarget == nil {
		return
	}

	monster.movingToTarget = nil
	monster.attackingTarget = nil
	monster.Attrs.SetStr("action", "idle")
}

func (monster *Monster) MovingTo(player *entity.Entity) {
	if monster.movingToTarget == player {
		// moving target not changed
		return
	}

	monster.movingToTarget = player
	monster.attackingTarget = nil
	monster.Attrs.SetStr("action", "move")
}

func (monster *Monster) Attacking(player *entity.Entity) {
	if monster.attackingTarget == player {
		return
	}

	monster.movingToTarget = nil
	monster.attackingTarget = player
	monster.Attrs.SetStr("action", "move")
}

func (monster *Monster) attack(player *Player) {
	monster.CallAllClients("DisplayAttack", player.ID)

	if player.GetInt("hp") <= 0 {
		return
	}

	player.TakeDamage(monster.GetDamage())
}

func (monster *Monster) GetDamage() int64 {
	return 10
}

func (monster *Monster) TakeDamage(damage int64) {
	hp := monster.GetInt("hp")
	hp = hp - damage
	if hp < 0 {
		hp = 0
	}

	monster.Attrs.SetInt("hp", hp)
	gwlog.Infof("%s TakeDamage %d => hp=%d", monster, damage, hp)
	if hp <= 0 {
		monster.Attrs.SetStr("action", "death")
		monster.Destroy()
	}
}
