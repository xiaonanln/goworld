package config

import (
	"testing"

	"github.com/xiaonanln/goworld/gwlog"
)

func TestLoadConfig(t *testing.T) {
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
