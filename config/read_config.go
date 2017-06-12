package config

import (
	"strings"

	"strconv"

	"fmt"

	"encoding/json"

	"sync"

	"github.com/xiaonanln/goworld/gwlog"
	"gopkg.in/ini.v1"
)

const (
	DEFAULT_CONFIG_FILE  = "goworld.ini"
	DEFAULT_LOCALHOST_IP = "127.0.0.1"
)

var (
	configFilePath = DEFAULT_CONFIG_FILE
	goWorldConfig  *GoWorldConfig
	configLock     sync.Mutex
)

type ServerConfig struct {
	Ip   string
	Port int
}

type DispatcherConfig struct {
	Ip   string
	Port int
}

type GoWorldConfig struct {
	dispatcher DispatcherConfig
	servers    map[int]*ServerConfig
	storage    StorageConfig
}

type StorageConfig struct {
	Type string
	// Filesystem Storage Configs
	Directory string // directory for filesystem storage
	// MongoDB storage configs
}

func SetConfigFile(f string) {
	configFilePath = f
}

func Get() *GoWorldConfig {
	configLock.Lock()
	defer configLock.Unlock() // protect concurrent access from Games & Gate
	if goWorldConfig == nil {
		goWorldConfig = readGoWorldConfig()
	}
	return goWorldConfig
}

func Reload() *GoWorldConfig {
	configLock.Lock()
	defer configLock.Unlock()

	goWorldConfig = nil
	return Get()
}

func GetServer(serverid int) *ServerConfig {
	return Get().servers[serverid]
}

func GetDispatcher() *DispatcherConfig {
	return &Get().dispatcher
}

func GetStorage() *StorageConfig {
	return &Get().storage
}

func DumpPretty(cfg interface{}) string {
	s, err := json.MarshalIndent(cfg, "", "    ")
	if err != nil {
		return err.Error()
	}
	return string(s)
}

func readGoWorldConfig() *GoWorldConfig {
	config := GoWorldConfig{
		servers: map[int]*ServerConfig{},
	}
	gwlog.Info("Using config file: %s", configFilePath)
	iniFile, err := ini.Load(configFilePath)
	checkConfigError(err, "")
	for _, sec := range iniFile.Sections() {
		secName := sec.Name()
		if secName == "DEFAULT" {
			continue
		}

		//gwlog.Info("Section %s", sec.Name())
		secName = strings.ToLower(secName)
		if secName == "dispatcher" {
			// dispatcher config
			readDispatcherConfig(sec, &config.dispatcher)
		} else if secName[:6] == "server" {
			// server config
			id, err := strconv.Atoi(secName[6:])
			checkConfigError(err, fmt.Sprintf("invalid server name: %s", secName))
			config.servers[id] = readServerConfig(sec)
		} else if secName == "storage" {
			// storage config
			readStorageConfig(sec, &config.storage)
		} else {
			gwlog.Warn("unknown section: %s", secName)
		}

	}
	return &config
}

func readServerConfig(sec *ini.Section) *ServerConfig {
	gc := &ServerConfig{
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

func readStorageConfig(sec *ini.Section, config *StorageConfig) {
	// setup default values
	config.Type = "filesystem"
	config.Directory = "_entity_storage"

	for _, key := range sec.Keys() {
		name := strings.ToLower(key.Name())
		if name == "type" {
			config.Type = key.MustString("filesystem")
		} else if name == "directory" {
			config.Directory = key.MustString("_entity_storage")
		}
	}

	validateStorageConfig(config)
}

func checkConfigError(err error, msg string) {
	if err != nil {
		if msg == "" {
			msg = err.Error()
		}
		gwlog.Panicf("read config error: %s", msg)
	}
}

func validateStorageConfig(config *StorageConfig) {
	if config.Type == "filesystem" {
		// directory must be set
	} else {
		gwlog.Panicf("unknown storage type: %s", config.Type)
	}
}
