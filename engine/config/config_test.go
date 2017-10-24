package config

import (
	"testing"

	"encoding/json"

	"fmt"

	"os"

	"github.com/bmizerany/assert"
	"github.com/xiaonanln/goworld/engine/gwlog"
)

func init() {
	SetConfigFile("../../goworld.ini.sample")
}

func TestLoad(t *testing.T) {
	config := Get()
	gwlog.Debugf("goworld config: \n%s", config)
	if config == nil {
		t.FailNow()
	}
	if config.Dispatcher.Ip == "" {
		t.Errorf("dispatch ip not found")
	}
	if config.Dispatcher.Port == 0 {
		t.Errorf("dispatcher port not found")
	}
	for gateid, gateConfig := range config.Gates {
		if gateConfig.Ip == "" {
			t.Errorf("gate %d ip not found", gateid)
		}
		if gateConfig.Port == 0 {
			t.Errorf("gate %d port not found", gateid)
		}
	}

	gwlog.Infof("read goworld config: %v", config)
}

func TestReload(t *testing.T) {
	Get()
	config := Reload()
	gwlog.Debugf("goworld config: \n%s", config)
}

func TestGetDispatcher(t *testing.T) {
	cfg := GetDispatcher()
	cfgStr, _ := json.Marshal(cfg)
	fmt.Printf("dispatcher config: %s", string(cfgStr))
}

func TestGetGame(t *testing.T) {
	for id := 1; id <= 10; id++ {
		cfg := GetGame(uint16(id))
		if cfg == nil {
			gwlog.Infof("Game %d not found", id)
		} else {
			gwlog.Infof("Game %d config: %v", id, cfg)
		}
	}
}

func TestGetStorage(t *testing.T) {
	cfg := GetStorage()
	if cfg == nil {
		t.Errorf("storage config not found")
	}
	gwlog.Infof("storage config:")
	fmt.Fprintf(os.Stderr, "%s\n", DumpPretty(cfg))
}

func TestGetKVDB(t *testing.T) {
	assert.T(t, GetKVDB() != nil, "kvdb config is nil")
}

func TestGetGameIDs(t *testing.T) {
	GetGameIDs()
}

func TestGetGate(t *testing.T) {
	GetGate(1)
}

func TestGetGateIDs(t *testing.T) {
	GetGateIDs()
	//assert.Equal(t, len(gids), 1, "gate num is wrong")
	//assert.Equal(t, gids[0], uint16(1), "gate id is not 1")
}

func TestSetConfigFile(t *testing.T) {
	SetConfigFile("goworld.ini")
}
