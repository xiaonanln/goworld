package game

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/xiaonanln/goworld/engine/entity"
	"github.com/xiaonanln/goworld/engine/gwlog"
)

func freezeFilename(gameid uint16) string {
	return fmt.Sprintf("game%d_freezed.dat", gameid)
}

func restoreFreezedEntities() error {
	t0 := time.Now()
	freezeFilename := freezeFilename(gameid)
	data, err := ioutil.ReadFile(freezeFilename)
	if err != nil {
		return err
	}

	t1 := time.Now()
	var freezeEntity entity.FreezeData
	freezePacker.UnpackMsg(data, &freezeEntity)
	t2 := time.Now()

	err = entity.RestoreFreezedEntities(&freezeEntity)
	t3 := time.Now()

	gwlog.Infof("Restored game service: load = %s, unpack = %s, restore = %s", t1.Sub(t0), t2.Sub(t1), t3.Sub(t2))
	return err
}
