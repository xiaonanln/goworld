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
	DEFAULT_LOG_LEVEL     = "debug"
	DEFAULT_STORAGE_DB    = "goworld"
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
	LogLevel     string
	GoMaxProcs   int
}

type DispatcherConfig struct {
	Ip        string
	Port      int
	LogFile   string
	LogStderr bool
	PProfIp   string
	PProfPort int
	LogLevel  string
}

type GoWorldConfig struct {
	Dispatcher   DispatcherConfig
	ServerCommon ServerConfig
	Servers      map[int]*ServerConfig
	Storage      StorageConfig
	KVDB         KVDBConfig
}

type StorageConfig struct {
	Type string
	// Filesystem Storage Configs
	Directory string // directory for filesystem storage
	// MongoDB storage configs
	Url string
	DB  string
}

type KVDBConfig struct {
	Type       string
	Url        string
	DB         string
	Collection string
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

func GetKVDB() *KVDBConfig {
	return &Get().KVDB
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
		} else if len(secName) > 6 && secName[:6] == "server" {
			// server config
			id, err := strconv.Atoi(secName[6:])
			checkConfigError(err, fmt.Sprintf("invalid server name: %s", secName))
			config.Servers[id] = readServerConfig(sec, &config.ServerCommon)
		} else if secName == "storage" {
			// storage config
			readStorageConfig(sec, &config.Storage)
		} else if secName == "kvdb" {
			// kvdb config
			readKVDBConfig(sec, &config.KVDB)
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
	scc.LogLevel = DEFAULT_LOG_LEVEL
	scc.SaveInterval = DEFAULT_SAVE_ITNERVAL
	scc.PProfIp = DEFAULT_PPROF_IP
	scc.PProfPort = 0 // pprof not enabled by default
	scc.GoMaxProcs = 0

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
		} else if name == "log_level" {
			sc.LogLevel = key.MustString(sc.LogLevel)
		} else if name == "gomaxprocs" {
			sc.GoMaxProcs = key.MustInt(sc.GoMaxProcs)
		} else {
			gwlog.Panicf("section %s has unknown key: %s", sec.Name(), key.Name())
		}
	}
}

func readDispatcherConfig(sec *ini.Section, config *DispatcherConfig) {
	config.Ip = DEFAULT_LOCALHOST_IP
	config.LogFile = ""
	config.LogStderr = true
	config.LogLevel = DEFAULT_LOG_LEVEL
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
		} else if name == "log_level" {
			config.LogLevel = key.MustString(config.LogLevel)
		} else {
			gwlog.Panicf("section %s has unknown key: %s", sec.Name(), key.Name())
		}
	}
	return
}

func readStorageConfig(sec *ini.Section, config *StorageConfig) {
	// setup default values
	config.Type = "filesystem"
	config.Directory = "_entity_storage"
	config.DB = DEFAULT_STORAGE_DB
	config.Url = ""

	for _, key := range sec.Keys() {
		name := strings.ToLower(key.Name())
		if name == "type" {
			config.Type = key.MustString(config.Type)
		} else if name == "directory" {
			config.Directory = key.MustString(config.Directory)
		} else if name == "url" {
			config.Url = key.MustString(config.Url)
		} else if name == "db" {
			config.DB = key.MustString(config.DB)
		} else {
			gwlog.Panicf("section %s has unknown key: %s", sec.Name(), key.Name())
		}
	}

	validateStorageConfig(config)
}

func readKVDBConfig(sec *ini.Section, config *KVDBConfig) {
	for _, key := range sec.Keys() {
		name := strings.ToLower(key.Name())
		if name == "type" {
			config.Type = key.MustString(config.Type)
		} else if name == "url" {
			config.Url = key.MustString(config.Url)
		} else if name == "db" {
			config.DB = key.MustString(config.DB)
		} else if name == "collection" {
			config.Collection = key.MustString(config.Collection)
		} else {
			gwlog.Panicf("section %s has unknown key: %s", sec.Name(), key.Name())
		}
	}

	validateKVDBConfig(config)
}

func validateKVDBConfig(config *KVDBConfig) {
	if config.Type == "" {
		// KVDB not enabled, it's OK
	} else if config.Type == "mongodb" {
		// must set DB and Collection for mongodb
		if config.Url == "" || config.DB == "" || config.Collection == "" {
			fmt.Fprintf(gwlog.GetOutput(), "%s\n", DumpPretty(config))
			gwlog.Panicf("invalid %s KVDB config above", config.Type)
		}
	} else {
		gwlog.Panicf("unknown storage type: %s", config.Type)
	}
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
		if config.Directory == "" {
			gwlog.Panicf("directory is not set in %s storage config", config.Type)
		}
	} else if config.Type == "mongodb" {
		if config.Url == "" {
			gwlog.Panicf("url is not set in %s storage config", config.Type)
		}
		if config.DB == "" {
			gwlog.Panicf("db is not set in %s storage config", config.Type)
		}
	} else {
		gwlog.Panicf("unknown storage type: %s", config.Type)
	}
}
