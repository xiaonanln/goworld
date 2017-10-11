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

	"github.com/go-ini/ini"
	"github.com/pkg/errors"
	"github.com/xiaonanln/goworld/engine/consts"
	"github.com/xiaonanln/goworld/engine/gwlog"
)

const (
	_DEFAULT_CONFIG_FILE   = "goworld.ini"
	_DEFAULT_LOCALHOST_IP  = "127.0.0.1"
	_DEFAULT_SAVE_ITNERVAL = time.Minute * 5
	_DEFAULT_HTTP_IP       = "127.0.0.1"
	_DEFAULT_LOG_LEVEL     = "debug"
	_DEFAULT_STORAGE_DB    = "goworld"
)

var (
	configFilePath = _DEFAULT_CONFIG_FILE
	goWorldConfig  *GoWorldConfig
	configLock     sync.Mutex
)

// GameConfig defines fields of game config
type GameConfig struct {
	BootEntity   string
	SaveInterval time.Duration
	LogFile      string
	LogStderr    bool
	HTTPIp       string
	HTTPPort     int
	LogLevel     string
	GoMaxProcs   int
}

// GateConfig defines fields of gate config
type GateConfig struct {
	Ip                 string
	Port               int
	LogFile            string
	LogStderr          bool
	HTTPIp             string
	HTTPPort           int
	LogLevel           string
	GoMaxProcs         int
	CompressConnection bool
	CompressFormat     string
}

// DispatcherConfig defines fields of dispatcher config
type DispatcherConfig struct {
	Ip        string
	Port      int
	LogFile   string
	LogStderr bool
	HTTPIp    string
	HTTPPort  int
	LogLevel  string
}

// GoWorldConfig defines the total GoWorld config file structure
type GoWorldConfig struct {
	Dispatcher DispatcherConfig
	GameCommon GameConfig
	GateCommon GateConfig
	Games      map[int]*GameConfig
	Gates      map[int]*GateConfig
	Storage    StorageConfig
	KVDB       KVDBConfig
}

// StorageConfig defines fields of storage config
type StorageConfig struct {
	Type      string // Type of storage (filesystem, mongodb, redis, mysql)
	Directory string // Directory of filesystem storage (filesystem)
	Url       string // Connection URL (mongodb, redis, mysql)
	DB        string // Database name (mongodb, redis)
	Driver    string // SQL Driver name (mysql)
}

// KVDBConfig defines fields of KVDB config
type KVDBConfig struct {
	Type       string
	Url        string // MongoDB
	DB         string // MongoDB
	Collection string // MongoDB
	Driver     string // SQL Driver: e.x. mysql

}

// SetConfigFile sets the config file path (goworld.ini by default)
func SetConfigFile(f string) {
	configFilePath = f
}

// Get returns the total GoWorld config
func Get() *GoWorldConfig {
	configLock.Lock()
	defer configLock.Unlock() // protect concurrent access from Games & Gate
	if goWorldConfig == nil {
		goWorldConfig = readGoWorldConfig()
	}
	return goWorldConfig
}

// Reload forces goworld server to reload the whole config
func Reload() *GoWorldConfig {
	configLock.Lock()
	goWorldConfig = nil
	configLock.Unlock()

	return Get()
}

// GetGame gets the game config of specified game ID
func GetGame(gameid uint16) *GameConfig {
	return Get().Games[int(gameid)]
}

// GetGate gets the gate config of specified gate ID
func GetGate(gateid uint16) *GateConfig {
	return Get().Gates[int(gateid)]
}

// GetGameIDs returns all game IDs
func GetGameIDs() []uint16 {
	cfg := Get()
	gameIDs := make([]int, 0, len(cfg.Games))
	for id := range cfg.Games {
		gameIDs = append(gameIDs, id)
	}
	sort.Ints(gameIDs)

	res := make([]uint16, len(gameIDs))
	for i, id := range gameIDs {
		res[i] = uint16(id)
	}
	return res
}

// GetGateIDs returns all gate IDs
func GetGateIDs() []uint16 {
	cfg := Get()
	gateIDs := make([]int, 0, len(cfg.Gates))
	for id := range cfg.Gates {
		gateIDs = append(gateIDs, id)
	}
	sort.Ints(gateIDs)

	res := make([]uint16, len(gateIDs))
	for i, id := range gateIDs {
		res[i] = uint16(id)
	}
	return res
}

// GetDispatcher returns the dispatcher config
func GetDispatcher() *DispatcherConfig {
	return &Get().Dispatcher
}

// GetStorage returns the storage config
func GetStorage() *StorageConfig {
	return &Get().Storage
}

// GetKVDB returns the KVDB config
func GetKVDB() *KVDBConfig {
	return &Get().KVDB
}

// DumpPretty format config to string in pretty format
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
	gwlog.Infof("Using config file: %s", configFilePath)
	iniFile, err := ini.Load(configFilePath)
	checkConfigError(err, "")
	gameCommonSec := iniFile.Section("game_common")
	readGameCommonConfig(gameCommonSec, &config.GameCommon)
	gateCommonSec := iniFile.Section("gate_common")
	readGateCommonConfig(gateCommonSec, &config.GateCommon)

	for _, sec := range iniFile.Sections() {
		secName := sec.Name()
		if secName == "DEFAULT" {
			continue
		}

		//gwlog.Infof("Section %s", sec.Name())
		secName = strings.ToLower(secName)
		if secName == "dispatcher" {
			// dispatcher config
			readDispatcherConfig(sec, &config.Dispatcher)
		} else if secName == "game_common" || secName == "gate_common" {
			// ignore common section here
		} else if len(secName) > 4 && secName[:4] == "game" {
			// game config
			id, err := strconv.Atoi(secName[4:])
			checkConfigError(err, fmt.Sprintf("invalid game name: %s", secName))
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
			gwlog.Errorf("unknown section: %s", secName)
		}

	}
	return &config
}

func readGameCommonConfig(section *ini.Section, scc *GameConfig) {
	scc.BootEntity = "Boot"
	scc.LogFile = "game.log"
	scc.LogStderr = true
	scc.LogLevel = _DEFAULT_LOG_LEVEL
	scc.SaveInterval = _DEFAULT_SAVE_ITNERVAL
	scc.HTTPIp = _DEFAULT_HTTP_IP
	scc.HTTPPort = 0 // pprof not enabled by default
	scc.GoMaxProcs = 0

	_readGameConfig(section, scc)
}

func readGameConfig(sec *ini.Section, gameCommonConfig *GameConfig) *GameConfig {
	var sc GameConfig = *gameCommonConfig // copy from game_common
	_readGameConfig(sec, &sc)
	// validate game config
	if sc.BootEntity == "" {
		panic("boot_entity is not set in game config")
	}
	return &sc
}

func _readGameConfig(sec *ini.Section, sc *GameConfig) {
	for _, key := range sec.Keys() {
		name := strings.ToLower(key.Name())
		if name == "boot_entity" {
			sc.BootEntity = key.MustString(sc.BootEntity)
		} else if name == "save_interval" {
			sc.SaveInterval = time.Second * time.Duration(key.MustInt(int(_DEFAULT_SAVE_ITNERVAL/time.Second)))
		} else if name == "log_file" {
			sc.LogFile = key.MustString(sc.LogFile)
		} else if name == "log_stderr" {
			sc.LogStderr = key.MustBool(sc.LogStderr)
		} else if name == "http_ip" {
			sc.HTTPIp = key.MustString(sc.HTTPIp)
		} else if name == "http_port" {
			sc.HTTPPort = key.MustInt(sc.HTTPPort)
		} else if name == "log_level" {
			sc.LogLevel = key.MustString(sc.LogLevel)
		} else if name == "gomaxprocs" {
			sc.GoMaxProcs = key.MustInt(sc.GoMaxProcs)
		} else {
			gwlog.Panicf("section %s has unknown key: %s", sec.Name(), key.Name())
		}
	}
}

func readGateCommonConfig(section *ini.Section, gcc *GateConfig) {
	gcc.LogFile = "gate.log"
	gcc.LogStderr = true
	gcc.LogLevel = _DEFAULT_LOG_LEVEL
	gcc.Ip = "0.0.0.0"
	gcc.HTTPIp = _DEFAULT_HTTP_IP
	gcc.HTTPPort = 0 // pprof not enabled by default
	gcc.GoMaxProcs = 0
	gcc.CompressFormat = ""

	_readGateConfig(section, gcc)
}

func readGateConfig(sec *ini.Section, gateCommonConfig *GateConfig) *GateConfig {
	var sc GateConfig = *gateCommonConfig // copy from game_common
	_readGateConfig(sec, &sc)
	// validate game config here
	if sc.CompressConnection && sc.CompressFormat == "" {
		gwlog.Fatalf("Gate %s: compress_connection is enabled, but compress format is empty", sec.Name())
	}
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
		} else if name == "http_ip" {
			sc.HTTPIp = key.MustString(sc.HTTPIp)
		} else if name == "http_port" {
			sc.HTTPPort = key.MustInt(sc.HTTPPort)
		} else if name == "log_level" {
			sc.LogLevel = key.MustString(sc.LogLevel)
		} else if name == "gomaxprocs" {
			sc.GoMaxProcs = key.MustInt(sc.GoMaxProcs)
		} else if name == "compress_connection" {
			sc.CompressConnection = key.MustBool(sc.CompressConnection)
		} else if name == "compress_format" {
			sc.CompressFormat = key.MustString(sc.CompressFormat)
		} else {
			gwlog.Panicf("section %s has unknown key: %s", sec.Name(), key.Name())
		}
	}
}

func readDispatcherConfig(sec *ini.Section, config *DispatcherConfig) {
	config.Ip = _DEFAULT_LOCALHOST_IP
	config.LogFile = ""
	config.LogStderr = true
	config.LogLevel = _DEFAULT_LOG_LEVEL
	config.HTTPIp = _DEFAULT_HTTP_IP
	config.HTTPPort = 0

	for _, key := range sec.Keys() {
		name := strings.ToLower(key.Name())
		if name == "ip" {
			config.Ip = key.MustString(_DEFAULT_LOCALHOST_IP)
		} else if name == "port" {
			config.Port = key.MustInt(0)
		} else if name == "log_file" {
			config.LogFile = key.MustString(config.LogFile)
		} else if name == "log_stderr" {
			config.LogStderr = key.MustBool(config.LogStderr)
		} else if name == "http_ip" {
			config.HTTPIp = key.MustString(config.HTTPIp)
		} else if name == "http_port" {
			config.HTTPPort = key.MustInt(config.HTTPPort)
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
	config.DB = _DEFAULT_STORAGE_DB
	config.Url = ""
	config.Driver = ""

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
		} else if name == "driver" {
			config.Driver = key.MustString(config.Driver)
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
		} else if name == "driver" {
			config.Driver = key.MustString(config.Driver)
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
		if config.Url == "" {
			fmt.Fprintf(gwlog.GetOutput(), "%s\n", DumpPretty(config))
			gwlog.Panicf("invalid %s KVDB config above", config.Type)
		}
		_, err := strconv.Atoi(config.DB) // make sure db is integer for redis
		if err != nil {
			gwlog.Panic(errors.Wrap(err, "redis db must be integer"))
		}
	} else if config.Type == "sql" {
		if config.Driver == "" {
			fmt.Fprintf(gwlog.GetOutput(), "%s\n", DumpPretty(config))
			gwlog.Panicf("invalid %s KVDB config above", config.Type)
		}
		if config.Url == "" {
			fmt.Fprintf(gwlog.GetOutput(), "%s\n", DumpPretty(config))
			gwlog.Panicf("invalid %s KVDB config above", config.Type)
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
		if config.Url == "" {
			gwlog.Panicf("redis host is not set")
		}
		if _, err := strconv.Atoi(config.DB); err != nil {
			gwlog.Panic(errors.Wrap(err, "redis db must be integer"))
		}
	} else if config.Type == "sql" {
		if config.Driver == "" {
			gwlog.Panicf("sql driver is not set")
		}
		if config.Url == "" {
			gwlog.Panicf("db url is not set")
		}

	} else {
		gwlog.Panicf("unknown storage type: %s", config.Type)
		if consts.DEBUG_MODE {
			os.Exit(2)
		}
	}
}
