package config

import (
	"testing"

	"encoding/json"

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

	for dispid, dispatcherConfig := range config._Dispatchers {
		if dispatcherConfig.AdvertiseAddr == "" {
			t.Errorf("dispatch %d: advertise addr not found", dispid)
		}
	}

	gwlog.Infof("read goworld config: %v", config)
}

func TestReload(t *testing.T) {
	Get()
	config := Reload()
	gwlog.Debugf("goworld config: \n%s", config)
}

func TestGetDeployment(t *testing.T) {
	cfg := GetDeployment()
	cfgStr, _ := json.Marshal(cfg)
	t.Logf("deployment config: %s", string(cfgStr))
}

func TestGetDispatcher(t *testing.T) {
	cfg := GetDispatcher(1)
	cfgStr, _ := json.Marshal(cfg)
	t.Logf("dispatcher config: %s", string(cfgStr))
}

func TestGetGame(t *testing.T) {
	for id := 1; id <= 10; id++ {
		cfg := GetGame(uint16(id))
		if cfg == nil {
			t.Logf("Game %d not found", id)
		} else {
			t.Logf("Game %d config: %v", id, cfg)
		}
	}
}

func TestGetStorage(t *testing.T) {
	cfg := GetStorage()
	if cfg == nil {
		t.Errorf("storage config not found")
	}
	gwlog.Infof("storage config:")
	t.Logf("%s\n", DumpPretty(cfg))
}

func TestGetKVDB(t *testing.T) {
	assert.T(t, GetKVDB() != nil, "kvdb config is nil")
}

func TestGetGate(t *testing.T) {
	GetGate(1)
}

func TestSetConfigFile(t *testing.T) {
	SetConfigFile("../../goworld.ini")
}
