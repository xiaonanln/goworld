package config

import (
	"strings"

	"github.com/xiaonanln/goworld/gwlog"
	"gopkg.in/ini.v1"
)

const (
	DEFAULT_CONFIG_FILENAME = "goworld.ini"
	DEFAULT_LOCALHOST_IP    = "127.0.0.1"
)

var (
	goWorldConfig *GoWorldConfig
)

type GameConfig struct {
	Ip   string
	Port int
}

type DispatcherConfig struct {
	Ip   string
	Port int
}

type GateConfig struct {
	Ip   string
	Port int
}

type GoWorldConfig struct {
	dispatcher DispatcherConfig
	games      map[string]*GameConfig
	gates      map[string]*GateConfig
}

func Get() *GoWorldConfig {
	if goWorldConfig == nil {
		goWorldConfig = readGoWorldConfig()
	}
	return goWorldConfig
}

func Reload() *GoWorldConfig {
	goWorldConfig = nil
	return Get()
}

func readGoWorldConfig() *GoWorldConfig {
	config := GoWorldConfig{
		games: map[string]*GameConfig{},
		gates: map[string]*GateConfig{},
	}
	iniFile, err := ini.Load(DEFAULT_CONFIG_FILENAME)
	checkConfigError(err)
	for _, sec := range iniFile.Sections() {
		secName := sec.Name()
		if secName == "DEFAULT" {
			continue
		}

		gwlog.Info("Section %s", sec.Name())
		if secName == "dispatcher" {
			// dispatcher config
			readDispatcherConfig(sec, &config.dispatcher)
		} else if secName[:4] == "game" {
			// game config
			config.games[secName] = readGameConfig(sec)
		} else if secName[:4] == "gate" {
			// gate config
			config.gates[secName] = readGateConfig(sec)
		} else {
			gwlog.Warn("unknown section: %s", secName)
		}

	}
	return &config
}

func readGameConfig(sec *ini.Section) *GameConfig {
	gc := &GameConfig{
		Ip: DEFAULT_LOCALHOST_IP,
	}
	for _, key := range sec.Keys() {
		name := strings.ToLower(key.Name())
		if name == "ip" {
			gc.Ip = key.MustString(DEFAULT_LOCALHOST_IP)
		} else if name == "port" {
			gc.Port = key.MustInt(0)
		}
	}
	return gc
}

func readGateConfig(sec *ini.Section) *GateConfig {
	gc := &GateConfig{
		Ip: DEFAULT_LOCALHOST_IP,
	}
	for _, key := range sec.Keys() {
		name := strings.ToLower(key.Name())
		if name == "ip" {
			gc.Ip = key.MustString(DEFAULT_LOCALHOST_IP)
		} else if name == "port" {
			gc.Port = key.MustInt(0)
		}
	}
	return gc
}

func readDispatcherConfig(sec *ini.Section, config *DispatcherConfig) {
	config.Ip = DEFAULT_LOCALHOST_IP
	for _, key := range sec.Keys() {
		name := strings.ToLower(key.Name())
		if name == "ip" {
			config.Ip = key.MustString(DEFAULT_LOCALHOST_IP)
		} else if name == "port" {
			config.Port = key.MustInt(0)
		}
	}
	return
}

func checkConfigError(err error) {
	if err != nil {
		gwlog.Panic(err)
	}
}
