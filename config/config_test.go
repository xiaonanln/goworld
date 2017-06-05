package config

import (
	"testing"

	"encoding/json"

	"fmt"

	"github.com/xiaonanln/goworld/gwlog"
)

func TestLoad(t *testing.T) {
	config := Get()
	gwlog.Debug("goworld config: \n%s", config)
	if config == nil {
		t.FailNow()
	}
	if config.dispatcher.Ip == "" {
		t.Errorf("dispatch ip not found")
	}
	if config.dispatcher.Port == 0 {
		t.Errorf("dispatcher port not found")
	}
	for gameName, gameConfig := range config.games {
		if gameConfig.Ip == "" {
			t.Errorf("game %s ip not found", gameName)
		}
		if gameConfig.Port == 0 {
			t.Errorf("game %s port not found", gameName)
		}
	}
	for gateName, gateConfig := range config.gates {
		if gateConfig.Ip == "" {
			t.Errorf("gate %s ip not found", gateName)
		}
		if gateConfig.Port == 0 {
			t.Errorf("gate %s port not found", gateName)
		}
	}
	gwlog.Info("read goworld config: %v", config)
}

func TestReload(t *testing.T) {
	config := Get()
	config = Reload()
	gwlog.Debug("goworld config: \n%s", config)
}

func TestGetDispatcher(t *testing.T) {
	cfg := GetDispatcher()
	cfgStr, _ := json.Marshal(cfg)
	fmt.Printf("dispatcher config: %s", string(cfgStr))
}

func TestGetGame(t *testing.T) {
	for id := 1; id <= 10; id++ {
		cfg := GetGame(id)
		if cfg == nil {
			gwlog.Info("Game %d not found", id)
		} else {
			gwlog.Info("Game %d config: %v", id, cfg)
		}
	}
}

func TestGetGate(t *testing.T) {

	for id := 1; id <= 5; id++ {
		cfg := GetGate(id)
		if cfg == nil {
			gwlog.Info("Gate %d not found", id)
		} else {
			gwlog.Info("Gate %d config: %v", id, cfg)
		}
	}
}
