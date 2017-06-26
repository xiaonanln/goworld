package config

import (
	"strings"

	"strconv"

	"fmt"

	"encoding/json"

	"sync"

	"sort"

	"time"

	"github.com/xiaonanln/goworld/gwlog"
	"gopkg.in/ini.v1"
)

const (
	DEFAULT_CONFIG_FILE   = "goworld.ini"
	DEFAULT_LOCALHOST_IP  = "127.0.0.1"
	DEFAULT_SAVE_ITNERVAL = time.Minute * 5
	DEFAULT_PPROF_IP      = "127.0.0.1"
)

var (
	configFilePath = DEFAULT_CONFIG_FILE
	goWorldConfig  *GoWorldConfig
	configLock     sync.Mutex
)

type ServerConfig struct {
	Ip           string
	Port         int
	BootEntity   string
	SaveInterval time.Duration
	LogFile      string
	LogStderr    bool
	PProfIp      string
	PProfPort    int
}

type DispatcherConfig struct {
	Ip        string
	Port      int
	LogFile   string
	LogStderr bool
	PProfIp   string
	PProfPort int
}

type GoWorldConfig struct {
	Dispatcher   DispatcherConfig
	ServerCommon ServerConfig
	Servers      map[int]*ServerConfig
	Storage      StorageConfig
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

func GetServer(serverid uint16) *ServerConfig {
	return Get().Servers[int(serverid)]
}

func GetServerIDs() []uint16 {
	cfg := Get()
	serverIDs := make([]int, 0, len(cfg.Servers))
	for id, _ := range cfg.Servers {
		serverIDs = append(serverIDs, id)
	}
	sort.Ints(serverIDs)

	res := make([]uint16, len(serverIDs))
	for i, id := range serverIDs {
		res[i] = uint16(id)
	}
	return res
}

func GetDispatcher() *DispatcherConfig {
	return &Get().Dispatcher
}

func GetStorage() *StorageConfig {
	return &Get().Storage
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
		Servers: map[int]*ServerConfig{},
	}
	gwlog.Info("Using config file: %s", configFilePath)
	iniFile, err := ini.Load(configFilePath)
	checkConfigError(err, "")
	serverCommonSec := iniFile.Section("server_common")
	readServerCommonConfig(serverCommonSec, &config.ServerCommon)

	for _, sec := range iniFile.Sections() {
		secName := sec.Name()
		if secName == "DEFAULT" {
			continue
		}

		//gwlog.Info("Section %s", sec.Name())
		secName = strings.ToLower(secName)
		if secName == "dispatcher" {
			// dispatcher config
			readDispatcherConfig(sec, &config.Dispatcher)
		} else if secName == "server_common" {
			// ignore server_common here
		} else if secName[:6] == "server" {
			// server config
			id, err := strconv.Atoi(secName[6:])
			checkConfigError(err, fmt.Sprintf("invalid server name: %s", secName))
			config.Servers[id] = readServerConfig(sec, &config.ServerCommon)
		} else if secName == "storage" {
			// storage config
			readStorageConfig(sec, &config.Storage)
		} else {
			gwlog.Error("unknown section: %s", secName)
		}

	}
	return &config
}

func readServerCommonConfig(section *ini.Section, scc *ServerConfig) {
	scc.BootEntity = "Boot"
	scc.Ip = "0.0.0.0"
	scc.LogFile = "server.log"
	scc.LogStderr = true
	scc.SaveInterval = DEFAULT_SAVE_ITNERVAL
	scc.PProfIp = DEFAULT_PPROF_IP
	scc.PProfPort = 0 // pprof not enabled by default

	_readServerConfig(section, scc)
}

func readServerConfig(sec *ini.Section, serverCommonConfig *ServerConfig) *ServerConfig {
	var sc ServerConfig = *serverCommonConfig // copy from server_common
	_readServerConfig(sec, &sc)
	// validate game config
	if sc.BootEntity == "" {
		panic("boot_entity is not set in server config")
	}
	return &sc
}

func _readServerConfig(sec *ini.Section, sc *ServerConfig) {
	for _, key := range sec.Keys() {
		name := strings.ToLower(key.Name())
		if name == "ip" {
			sc.Ip = key.MustString(sc.Ip)
		} else if name == "port" {
			sc.Port = key.MustInt(sc.Port)
		} else if name == "boot_entity" {
			sc.BootEntity = key.MustString(sc.BootEntity)
		} else if name == "save_interval" {
			sc.SaveInterval = time.Second * time.Duration(key.MustInt(int(DEFAULT_SAVE_ITNERVAL/time.Second)))
		} else if name == "log_file" {
			sc.LogFile = key.MustString(sc.LogFile)
		} else if name == "log_stderr" {
			sc.LogStderr = key.MustBool(sc.LogStderr)
		} else if name == "pprof_ip" {
			sc.PProfIp = key.MustString(sc.PProfIp)
		} else if name == "pprof_port" {
			sc.PProfPort = key.MustInt(sc.PProfPort)
		}
	}
}

func readDispatcherConfig(sec *ini.Section, config *DispatcherConfig) {
	config.Ip = DEFAULT_LOCALHOST_IP
	config.LogFile = ""
	config.LogStderr = true
	config.PProfIp = DEFAULT_PPROF_IP
	config.PProfPort = 0

	for _, key := range sec.Keys() {
		name := strings.ToLower(key.Name())
		if name == "ip" {
			config.Ip = key.MustString(DEFAULT_LOCALHOST_IP)
		} else if name == "port" {
			config.Port = key.MustInt(0)
		} else if name == "log_file" {
			config.LogFile = key.MustString(config.LogFile)
		} else if name == "log_stderr" {
			config.LogStderr = key.MustBool(config.LogStderr)
		} else if name == "pprof_ip" {
			config.PProfIp = key.MustString(config.PProfIp)
		} else if name == "pprof_port" {
			config.PProfPort = key.MustInt(config.PProfPort)
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
