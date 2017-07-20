package config

import (
	"strings"

	"strconv"

	"fmt"

	"encoding/json"

	"sync"

	"sort"

	"time"

	"os"

	"github.com/pkg/errors"
	"github.com/xiaonanln/goworld/consts"
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

type GameConfig struct {
	BootEntity   string
	SaveInterval time.Duration
	LogFile      string
	LogStderr    bool
	PProfIp      string
	PProfPort    int
	LogLevel     string
	GoMaxProcs   int
}

type GateConfig struct {
	Ip                 string
	Port               int
	LogFile            string
	LogStderr          bool
	PProfIp            string
	PProfPort          int
	LogLevel           string
	GoMaxProcs         int
	CompressConnection bool
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
	Dispatcher DispatcherConfig
	GameCommon GameConfig
	GateCommon GateConfig
	Games      map[int]*GameConfig
	Gates      map[int]*GateConfig
	Storage    StorageConfig
	KVDB       KVDBConfig
}

type StorageConfig struct {
	Type string
	// Filesystem Storage Configs
	Directory string // directory for filesystem storage
	// MongoDB storage configs
	Url  string
	DB   string
	Host string // Redis host
}

type KVDBConfig struct {
	Type       string
	Url        string // MongoDB
	DB         string // MongoDB
	Collection string // MongoDB
	Host       string // Redis

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

func GetGame(serverid uint16) *GameConfig {
	return Get().Games[int(serverid)]
}

func GetGate(gateid uint16) *GateConfig {
	return Get().Gates[int(gateid)]
}

func GetGameIDs() []uint16 {
	cfg := Get()
	serverIDs := make([]int, 0, len(cfg.Games))
	for id, _ := range cfg.Games {
		serverIDs = append(serverIDs, id)
	}
	sort.Ints(serverIDs)

	res := make([]uint16, len(serverIDs))
	for i, id := range serverIDs {
		res[i] = uint16(id)
	}
	return res
}

func GetGateIDs() []uint16 {
	cfg := Get()
	gateIDs := make([]int, 0, len(cfg.Gates))
	for id, _ := range cfg.Gates {
		gateIDs = append(gateIDs, id)
	}
	sort.Ints(gateIDs)

	res := make([]uint16, len(gateIDs))
	for i, id := range gateIDs {
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
		Games: map[int]*GameConfig{},
		Gates: map[int]*GateConfig{},
	}
	gwlog.Info("Using config file: %s", configFilePath)
	iniFile, err := ini.Load(configFilePath)
	checkConfigError(err, "")
	serverCommonSec := iniFile.Section("server_common")
	readGameCommonConfig(serverCommonSec, &config.GameCommon)
	gateCommonSec := iniFile.Section("gate_common")
	readGateCommonConfig(gateCommonSec, &config.GateCommon)

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
		} else if secName == "server_common" || secName == "gate_common" {
			// ignore common section here
		} else if len(secName) > 6 && secName[:6] == "server" {
			// server config
			id, err := strconv.Atoi(secName[6:])
			checkConfigError(err, fmt.Sprintf("invalid server name: %s", secName))
			config.Games[id] = readGameConfig(sec, &config.GameCommon)
		} else if len(secName) > 4 && secName[:4] == "gate" {
			id, err := strconv.Atoi(secName[4:])
			checkConfigError(err, fmt.Sprintf("invalid gate name: %s", secName))
			config.Gates[id] = readGateConfig(sec, &config.GateCommon)
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

func readGameCommonConfig(section *ini.Section, scc *GameConfig) {
	scc.BootEntity = "Boot"
	scc.LogFile = "server.log"
	scc.LogStderr = true
	scc.LogLevel = DEFAULT_LOG_LEVEL
	scc.SaveInterval = DEFAULT_SAVE_ITNERVAL
	scc.PProfIp = DEFAULT_PPROF_IP
	scc.PProfPort = 0 // pprof not enabled by default
	scc.GoMaxProcs = 0

	_readGameConfig(section, scc)
}

func readGameConfig(sec *ini.Section, serverCommonConfig *GameConfig) *GameConfig {
	var sc GameConfig = *serverCommonConfig // copy from server_common
	_readGameConfig(sec, &sc)
	// validate game config
	if sc.BootEntity == "" {
		panic("boot_entity is not set in server config")
	}
	return &sc
}

func _readGameConfig(sec *ini.Section, sc *GameConfig) {
	for _, key := range sec.Keys() {
		name := strings.ToLower(key.Name())
		if name == "boot_entity" {
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

func readGateCommonConfig(section *ini.Section, scc *GateConfig) {
	scc.LogFile = "gate.log"
	scc.LogStderr = true
	scc.LogLevel = DEFAULT_LOG_LEVEL
	scc.PProfIp = DEFAULT_PPROF_IP
	scc.PProfPort = 0 // pprof not enabled by default
	scc.GoMaxProcs = 0

	_readGateConfig(section, scc)
}

func readGateConfig(sec *ini.Section, gateCommonConfig *GateConfig) *GateConfig {
	var sc GateConfig = *gateCommonConfig // copy from server_common
	_readGateConfig(sec, &sc)
	// validate game config
	return &sc
}

func _readGateConfig(sec *ini.Section, sc *GateConfig) {
	for _, key := range sec.Keys() {
		name := strings.ToLower(key.Name())
		if name == "ip" {
			sc.Ip = key.MustString(sc.Ip)
		} else if name == "port" {
			sc.Port = key.MustInt(sc.Port)
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
		} else if name == "compress_connection" {
			sc.CompressConnection = key.MustBool(sc.CompressConnection)
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
		} else if name == "host" {
			config.Host = key.MustString(config.Host)
		} else {
			gwlog.Panicf("section %s has unknown key: %s", sec.Name(), key.Name())
		}
	}

	if config.Type == "redis" {
		if config.DB == "" {
			config.DB = "0"
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
		} else if name == "host" {
			config.Host = key.MustString(config.Host)
		} else {
			gwlog.Panicf("section %s has unknown key: %s", sec.Name(), key.Name())
		}
	}

	if config.Type == "redis" {
		if config.DB == "" {
			config.DB = "0"
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
	} else if config.Type == "redis" {
		if config.Host == "" {
			fmt.Fprintf(gwlog.GetOutput(), "%s\n", DumpPretty(config))
			gwlog.Panicf("invalid %s KVDB config above", config.Type)
		}
		_, err := strconv.Atoi(config.DB) // make sure db is integer for redis
		if err != nil {
			gwlog.Panic(errors.Wrap(err, "redis db must be integer"))
		}
	} else {
		gwlog.Panicf("unknown storage type: %s", config.Type)
		if consts.DEBUG_MODE {
			os.Exit(2)
		}
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
	} else if config.Type == "redis" {
		if config.Host == "" {
			gwlog.Panicf("redis host is not set")
		}
		if _, err := strconv.Atoi(config.DB); err != nil {
			gwlog.Panic(errors.Wrap(err, "redis db must be integer"))
		}
	} else {
		gwlog.Panicf("unknown storage type: %s", config.Type)
		if consts.DEBUG_MODE {
			os.Exit(2)
		}
	}
}
