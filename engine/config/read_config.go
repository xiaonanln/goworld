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

	"path"

	"github.com/go-ini/ini"
	"github.com/pkg/errors"
	"github.com/xiaonanln/goworld/engine/common"
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
	BootEntity             string
	SaveInterval           time.Duration
	LogFile                string
	LogStderr              bool
	HTTPIp                 string
	HTTPPort               int
	LogLevel               string
	GoMaxProcs             int
	PositionSyncIntervalMS int
	BanBootEntity          bool
}

// GateConfig defines fields of gate config
type GateConfig struct {
	Ip                     string
	Port                   int
	LogFile                string
	LogStderr              bool
	HTTPIp                 string
	HTTPPort               int
	LogLevel               string
	GoMaxProcs             int
	CompressConnection     bool
	CompressFormat         string
	EncryptConnection      bool
	RSAKey                 string
	RSACertificate         string
	HeartbeatCheckInterval int
	PositionSyncIntervalMS int
}

// DispatcherConfig defines fields of dispatcher config
type DispatcherConfig struct {
	BindIp    string
	BindPort  int
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
	DispatcherCommon DispatcherConfig
	GameCommon       GameConfig
	GateCommon       GateConfig
	Dispatchers      map[int]*DispatcherConfig
	Games            map[int]*GameConfig
	Gates            map[int]*GateConfig
	Storage          StorageConfig
	KVDB             KVDBConfig
}

// StorageConfig defines fields of storage config
type StorageConfig struct {
	Type       string // Type of storage (filesystem, mongodb, redis, mysql)
	Directory  string // Directory of filesystem storage (filesystem)
	Url        string // Connection URL (mongodb, redis, mysql)
	DB         string // Database name (mongodb, redis)
	Driver     string // SQL Driver name (mysql)
	StartNodes common.StringSet
}

// KVDBConfig defines fields of KVDB config
type KVDBConfig struct {
	Type       string
	Url        string // MongoDB
	DB         string // MongoDB
	Collection string // MongoDB
	Driver     string // SQL Driver: e.x. mysql
	StartNodes common.StringSet
}

// SetConfigFile sets the config file path (goworld.ini by default)
func SetConfigFile(f string) {
	configFilePath = f
}

// GetConfigDir returns the directory of goworld.ini
func GetConfigDir() string {
	dir, _ := path.Split(configFilePath)
	return dir
}

// GetConfigFilePath returns the config file path
func GetConfigFilePath() string {
	return configFilePath
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

// GetDispatcherIDs returns all dispatcher IDs
func GetDispatcherIDs() []uint16 {
	cfg := Get()
	dispIDs := make([]int, 0, len(cfg.Dispatchers))
	for id := range cfg.Dispatchers {
		dispIDs = append(dispIDs, id)
	}
	sort.Ints(dispIDs)

	res := make([]uint16, len(dispIDs))
	for i, id := range dispIDs {
		res[i] = uint16(id)
	}
	return res
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
func GetDispatcher(dispid uint16) *DispatcherConfig {
	return Get().Dispatchers[int(dispid)]
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
		Dispatchers: map[int]*DispatcherConfig{},
		Games:       map[int]*GameConfig{},
		Gates:       map[int]*GateConfig{},
	}
	gwlog.Infof("Using config file: %s", configFilePath)
	iniFile, err := ini.Load(configFilePath)
	checkConfigError(err, "")
	gameCommonSec := iniFile.Section("game_common")
	readGameCommonConfig(gameCommonSec, &config.GameCommon)
	gateCommonSec := iniFile.Section("gate_common")
	readGateCommonConfig(gateCommonSec, &config.GateCommon)
	dispatcherCommonSec := iniFile.Section("dispatcher_common")
	readDispatcherCommonConfig(dispatcherCommonSec, &config.DispatcherCommon)

	for _, sec := range iniFile.Sections() {
		secName := sec.Name()
		if secName == "DEFAULT" {
			continue
		}

		//gwlog.Infof("Section %s", sec.Name())
		secName = strings.ToLower(secName)
		if secName == "game_common" || secName == "gate_common" || secName == "dispatcher_common" {
			// ignore common section here
		} else if len(secName) > 10 && secName[:10] == "dispatcher" {
			// dispatcher config
			id, err := strconv.Atoi(secName[10:])
			checkConfigError(err, fmt.Sprintf("invalid dispatcher name: %s", secName))
			config.Dispatchers[id] = readDispatcherConfig(sec, &config.DispatcherCommon)
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

	validateConfig(&config)
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
	scc.PositionSyncIntervalMS = 100 // sync positions per 100ms by default

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
		} else if name == "position_sync_interval_ms" {
			sc.PositionSyncIntervalMS = key.MustInt(sc.PositionSyncIntervalMS)
		} else if name == "ban_boot_entity" {
			sc.BanBootEntity = key.MustBool(sc.BanBootEntity)
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
	gcc.CompressFormat = "gwsnappy"
	gcc.RSAKey = "rsa.key"
	gcc.RSACertificate = "rsa.crt"
	gcc.HeartbeatCheckInterval = 0
	gcc.PositionSyncIntervalMS = 100

	_readGateConfig(section, gcc)
}

func readGateConfig(sec *ini.Section, gateCommonConfig *GateConfig) *GateConfig {
	var sc GateConfig = *gateCommonConfig // copy from game_common
	_readGateConfig(sec, &sc)
	// validate game config here
	if sc.CompressConnection && sc.CompressFormat == "" {
		gwlog.Fatalf("Gate %s: compress_connection is enabled, but compress format is not set", sec.Name())
	}
	if sc.EncryptConnection && sc.RSAKey == "" {
		gwlog.Fatalf("Gate %s: encrypt_connection is enabled, but rsa_key is not set", sec.Name())
	}
	if sc.EncryptConnection && sc.RSACertificate == "" {
		gwlog.Fatalf("Gate %s: encrypt_connection is enabled, but rsa_certificate is not set", sec.Name())
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
		} else if name == "encrypt_connection" {
			sc.EncryptConnection = key.MustBool(sc.EncryptConnection)
		} else if name == "rsa_key" {
			sc.RSAKey = key.MustString(sc.RSAKey)
		} else if name == "rsa_certificate" {
			sc.RSACertificate = key.MustString(sc.RSACertificate)
		} else if name == "heartbeat_check_interval" {
			sc.HeartbeatCheckInterval = key.MustInt(sc.HeartbeatCheckInterval)
		} else if name == "position_sync_interval_ms" {
			sc.PositionSyncIntervalMS = key.MustInt(sc.PositionSyncIntervalMS)
		} else {
			gwlog.Panicf("section %s has unknown key: %s", sec.Name(), key.Name())
		}
	}
}

func readDispatcherCommonConfig(section *ini.Section, dc *DispatcherConfig) {
	dc.BindIp = _DEFAULT_LOCALHOST_IP
	dc.Ip = _DEFAULT_LOCALHOST_IP
	dc.LogFile = "dispatcher.log"
	dc.LogStderr = true
	dc.LogLevel = _DEFAULT_LOG_LEVEL
	dc.HTTPIp = _DEFAULT_HTTP_IP
	dc.HTTPPort = 0

	_readDispatcherConfig(section, dc)
}

func readDispatcherConfig(sec *ini.Section, dispatcherCommonConfig *DispatcherConfig) *DispatcherConfig {
	dc := *dispatcherCommonConfig // copy from game_common
	_readDispatcherConfig(sec, &dc)
	// validate dispatcher config
	return &dc
}

func _readDispatcherConfig(sec *ini.Section, config *DispatcherConfig) {
	for _, key := range sec.Keys() {
		name := strings.ToLower(key.Name())
		if name == "ip" {
			config.Ip = key.MustString(config.Ip)
		} else if name == "port" {
			config.Port = key.MustInt(config.Port)
		} else if name == "bind_ip" {
			config.BindIp = key.MustString(config.BindIp)
		} else if name == "bind_port" {
			config.BindPort = key.MustInt(config.Port)
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
	config.StartNodes = common.StringSet{}

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
		} else if strings.HasPrefix(name, "start_nodes_") {
			config.StartNodes.Add(key.MustString(""))
		} else {
			gwlog.Fatalf("section %s has unknown key: %s", sec.Name(), key.Name())
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
	config.StartNodes = common.StringSet{}
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
		} else if strings.HasPrefix(name, "start_nodes_") {
			config.StartNodes.Add(key.MustString(""))
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
	} else if config.Type == "redis_cluster" {
		if len(config.StartNodes) == 0 {
			gwlog.Panicf("must have at least 1 start_nodes for [kvdb].redis_cluster")
		}
		for s := range config.StartNodes {
			if s == "" {
				gwlog.Panicf("start_nodes must not be empty")
			}
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
	} else if config.Type == "redis_cluster" {
		if len(config.StartNodes) == 0 {
			gwlog.Panicf("must have at least 1 start_nodes for [storage].redis_cluster")
		}
		for s := range config.StartNodes {
			if s == "" {
				gwlog.Panicf("start_nodes must not be empty")
			}
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

func validateConfig(config *GoWorldConfig) {

	dispatchersNum := len(config.Dispatchers)
	if dispatchersNum <= 0 {
		gwlog.Panicf("dispatcher not found in config file, must has at least 1 dispatcher")
	}

	for dispatcherid := 1; dispatcherid <= dispatchersNum; dispatcherid++ {
		if _, ok := config.Dispatchers[dispatcherid]; !ok {
			gwlog.Panicf("found %d dispatchers in config file, but dispatcher%d is not found. dispatcherid must be 1~%d", dispatchersNum, dispatcherid, dispatchersNum)
		}
	}

	gamesNum := len(config.Games)
	if gamesNum <= 0 {
		gwlog.Panicf("game not found in config file, must has at least 1 game")
	}

	hasNotBanBootEntityGame := false
	for gameid := 1; gameid <= gamesNum; gameid++ {
		if gameCfg, ok := config.Games[gameid]; ok {
			if !gameCfg.BanBootEntity {
				hasNotBanBootEntityGame = true //
			}
		} else {
			gwlog.Panicf("found %d games in config file, but game%d is not found. gameid must be 1~%d", gamesNum, gameid, gamesNum)
		}
	}
	if !hasNotBanBootEntityGame {
		gwlog.Panicf("must has at least 1 game with ban_boot_entity = false!")
	}

	gatesNum := len(config.Gates)
	if gatesNum <= 0 {
		gwlog.Panicf("gate not found in config file, must has at least 1 gate")
	}

	for gateid := 1; gateid <= gatesNum; gateid++ {
		if _, ok := config.Gates[gateid]; !ok {
			gwlog.Panicf("found %d gates in config file, but gate%d is not found. gateid must be 1~%d", gatesNum, gateid, gatesNum)
		}
	}
}
