package config

import (
	"strings"

	"strconv"

	"fmt"

	"encoding/json"

	"sync"

	"sort"

	"time"

	"path"

	"github.com/go-ini/ini"
	"github.com/pkg/errors"
	"github.com/xiaonanln/goworld/engine/common"
	"github.com/xiaonanln/goworld/engine/gwlog"
)

const (
	_DEFAULT_CONFIG_FILE   = "goworld.ini"
	_DEFAULT_SAVE_ITNERVAL = time.Minute * 5
	_DEFAULT_LOG_LEVEL     = "debug"
	_DEFAULT_STORAGE_DB    = "goworld"
)

var (
	configFilePath = _DEFAULT_CONFIG_FILE
	goWorldConfig  *GoWorldConfig
	configLock     sync.Mutex
)

// DeploymentConfig defines fields of deployment config
type DeploymentConfig struct {
	DesiredDispatchers int `ini:"desired_dispatchers"`
	DesiredGames       int `ini:"desired_games"`
	DesiredGates       int `ini:"desired_gates"`
}

// GameConfig defines fields of game config
type GameConfig struct {
	BootEntity             string
	SaveInterval           time.Duration
	LogFile                string
	LogStderr              bool
	HTTPAddr               string
	LogLevel               string
	GoMaxProcs             int
	PositionSyncIntervalMS int
	BanBootEntity          bool
}

// GateConfig defines fields of gate config
type GateConfig struct {
	ListenAddr             string
	LogFile                string
	LogStderr              bool
	HTTPAddr               string
	LogLevel               string
	GoMaxProcs             int
	CompressConnection     bool
	EncryptConnection      bool
	RSAKey                 string
	RSACertificate         string
	HeartbeatCheckInterval int
	PositionSyncIntervalMS int
}

// DispatcherConfig defines fields of dispatcher config
type DispatcherConfig struct {
	ListenAddr    string
	AdvertiseAddr string
	HTTPAddr      string
	LogFile       string
	LogStderr     bool
	LogLevel      string
}

// GoWorldConfig defines the total GoWorld config file structure
type GoWorldConfig struct {
	Deployment       DeploymentConfig
	DispatcherCommon DispatcherConfig
	GameCommon       GameConfig
	GateCommon       GateConfig
	_Dispatchers     map[uint16]*DispatcherConfig
	_Games           map[uint16]*GameConfig
	_Gates           map[uint16]*GateConfig
	Storage          StorageConfig
	KVDB             KVDBConfig
	Debug            DebugConfig
}

// StorageConfig defines fields of storage config
type StorageConfig struct {
	Type       string // Type of storage (mongodb)
	Url        string // Connection URL (mongodb)
	DB         string // Database name (mongodb)
	StartNodes common.StringSet
}

// KVDBConfig defines fields of KVDB config
type KVDBConfig struct {
	Type       string
	Url        string // MongoDB
	DB         string // MongoDB
	Collection string // MongoDB
	StartNodes common.StringSet
}

type DebugConfig struct {
	Debug bool
}

// SetConfigFile sets the config file path (goworld.ini by default)
func SetConfigFile(f string) {
	configLock.Lock()
	if configFilePath == f {
		configLock.Unlock()
		return
	}

	configFilePath = f
	configLock.Unlock()

	Reload()
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
		gwlog.Infof(">>> config <<< debug = %v", goWorldConfig.Debug.Debug)
		gwlog.Infof(">>> config <<< desired dispatcher count = %d", goWorldConfig.Deployment.DesiredDispatchers)
		gwlog.Infof(">>> config <<< desired game count = %d", goWorldConfig.Deployment.DesiredGames)
		gwlog.Infof(">>> config <<< desired gate count = %d", goWorldConfig.Deployment.DesiredGates)
		gwlog.Infof(">>> config <<< storage type = %s", goWorldConfig.Storage.Type)
		gwlog.Infof(">>> config <<< KVDB type = %s", goWorldConfig.KVDB.Type)
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

func GetDeployment() *DeploymentConfig {
	return &Get().Deployment
}

// GetGame gets the game config of specified game ID
func GetGame(gameid uint16) *GameConfig {
	cfg := Get()._Games[gameid]
	if cfg == nil {
		cfg = &Get().GameCommon
	}
	return cfg
}

// GetGate gets the gate config of specified gate ID
func GetGate(gateid uint16) *GateConfig {
	cfg := Get()._Gates[gateid]
	if cfg == nil {
		cfg = &Get().GateCommon
	}
	return cfg
}

// GetDispatcherIDs returns all dispatcher IDs
func GetDispatcherIDs() []uint16 {
	cfg := Get()
	dispIDs := make([]int, 0, len(cfg._Dispatchers))
	for id := range cfg._Dispatchers {
		dispIDs = append(dispIDs, int(id))
	}
	sort.Ints(dispIDs)

	res := make([]uint16, len(dispIDs))
	for i, id := range dispIDs {
		res[i] = uint16(id)
	}
	return res
}

// GetDispatcher returns the dispatcher config
func GetDispatcher(dispid uint16) *DispatcherConfig {
	return Get()._Dispatchers[dispid]
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

func Debug() bool {
	return Get().Debug.Debug
}

func readGoWorldConfig() *GoWorldConfig {
	config := GoWorldConfig{
		_Dispatchers: map[uint16]*DispatcherConfig{},
		_Games:       map[uint16]*GameConfig{},
		_Gates:       map[uint16]*GateConfig{},
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
	deploymentSec := iniFile.Section("deployment")
	if deploymentSec == nil {
		gwlog.Fatalf("[deployment] section not found in config file")
	}
	readDeploymentConfig(deploymentSec, &config.Deployment)
	for _, sec := range iniFile.Sections() {
		secName := sec.Name()
		if secName == "DEFAULT" {
			continue
		}

		//gwlog.Infof("Section %s", sec.Name())
		secName = strings.ToLower(secName)
		if secName == "game_common" || secName == "gate_common" || secName == "dispatcher_common" {
			// ignore common section here
		} else if secName == "deployment" {
			// deployment section already read
		} else if len(secName) > 10 && secName[:10] == "dispatcher" {
			// dispatcher config
			id, err := strconv.Atoi(secName[10:])
			checkConfigError(err, fmt.Sprintf("invalid dispatcher name: %s", secName))
			if id > config.Deployment.DesiredDispatchers {
				gwlog.Warnf("Section [%s] is ignored because [deployment].desired_dispatchers = %d", secName, config.Deployment.DesiredDispatchers)
				continue
			}

			config._Dispatchers[uint16(id)] = readDispatcherConfig(sec, &config.DispatcherCommon)
		} else if len(secName) > 4 && secName[:4] == "game" {
			// game config
			id, err := strconv.Atoi(secName[4:])
			checkConfigError(err, fmt.Sprintf("invalid game name: %s", secName))
			config._Games[uint16(id)] = readGameConfig(sec, &config.GameCommon)
		} else if len(secName) > 4 && secName[:4] == "gate" {
			id, err := strconv.Atoi(secName[4:])
			checkConfigError(err, fmt.Sprintf("invalid gate name: %s", secName))
			config._Gates[uint16(id)] = readGateConfig(sec, &config.GateCommon)
		} else if secName == "storage" {
			// storage config
			readStorageConfig(sec, &config.Storage)
		} else if secName == "kvdb" {
			// kvdb config
			readKVDBConfig(sec, &config.KVDB)
		} else if secName == "debug" {
			// debug config
			readDebugConfig(sec, &config.Debug)
		} else {
			gwlog.Fatalf("unknown section: %s", secName)
		}

	}

	for dispid := uint16(1); int(dispid) <= config.Deployment.DesiredDispatchers; dispid++ {
		if _, ok := config._Dispatchers[dispid]; !ok {
			// dispatcher config not found, fix it
			config._Dispatchers[dispid] = &config.DispatcherCommon
		}
	}

	validateConfig(&config)
	return &config
}

func readDeploymentConfig(sec *ini.Section, config *DeploymentConfig) {
	sec.MapTo(config)
}

func readGameCommonConfig(section *ini.Section, scc *GameConfig) {
	scc.BootEntity = "Boot"
	scc.LogFile = "game.log"
	scc.LogStderr = true
	scc.LogLevel = _DEFAULT_LOG_LEVEL
	scc.SaveInterval = _DEFAULT_SAVE_ITNERVAL
	scc.HTTPAddr = "127.0.0.1:25000"
	scc.GoMaxProcs = 0
	scc.PositionSyncIntervalMS = 100 // sync positions per 100ms by default

	_readGameConfig(section, scc)
}

func readGameConfig(sec *ini.Section, gameCommonConfig *GameConfig) *GameConfig {
	var sc = *gameCommonConfig // copy from game_common
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
		} else if name == "http_addr" {
			sc.HTTPAddr = key.MustString(sc.HTTPAddr)
		} else if name == "log_level" {
			sc.LogLevel = key.MustString(sc.LogLevel)
		} else if name == "gomaxprocs" {
			sc.GoMaxProcs = key.MustInt(sc.GoMaxProcs)
		} else if name == "position_sync_interval_ms" {
			sc.PositionSyncIntervalMS = key.MustInt(sc.PositionSyncIntervalMS)
		} else if name == "ban_boot_entity" {
			sc.BanBootEntity = key.MustBool(sc.BanBootEntity)
		} else {
			gwlog.Fatalf("section %s has unknown key: %s", sec.Name(), key.Name())
		}
	}
}

func readGateCommonConfig(section *ini.Section, gcc *GateConfig) {
	gcc.LogFile = "gate.log"
	gcc.LogStderr = true
	gcc.LogLevel = _DEFAULT_LOG_LEVEL
	gcc.ListenAddr = "0.0.0.0:14000"
	gcc.HTTPAddr = "127.0.0.1:24000"
	gcc.GoMaxProcs = 0
	gcc.RSAKey = "rsa.key"
	gcc.RSACertificate = "rsa.crt"
	gcc.HeartbeatCheckInterval = 0
	gcc.PositionSyncIntervalMS = 100

	_readGateConfig(section, gcc)
}

func readGateConfig(sec *ini.Section, gateCommonConfig *GateConfig) *GateConfig {
	var sc = *gateCommonConfig // copy from game_common
	_readGateConfig(sec, &sc)
	// validate game config here
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
		if name == "listen_addr" {
			sc.ListenAddr = key.MustString(sc.ListenAddr)
		} else if name == "log_file" {
			sc.LogFile = key.MustString(sc.LogFile)
		} else if name == "log_stderr" {
			sc.LogStderr = key.MustBool(sc.LogStderr)
		} else if name == "http_addr" {
			sc.HTTPAddr = key.MustString(sc.HTTPAddr)
		} else if name == "log_level" {
			sc.LogLevel = key.MustString(sc.LogLevel)
		} else if name == "gomaxprocs" {
			sc.GoMaxProcs = key.MustInt(sc.GoMaxProcs)
		} else if name == "compress_connection" {
			sc.CompressConnection = key.MustBool(sc.CompressConnection)
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
			gwlog.Fatalf("section %s has unknown key: %s", sec.Name(), key.Name())
		}
	}
}

func readDispatcherCommonConfig(section *ini.Section, dc *DispatcherConfig) {
	dc.ListenAddr = "127.0.0.1:13000"
	dc.AdvertiseAddr = "127.0.0.1:13000"
	dc.HTTPAddr = "127.0.0.1:23000"
	dc.LogFile = "dispatcher.log"
	dc.LogStderr = true
	dc.LogLevel = _DEFAULT_LOG_LEVEL

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
		if name == "advertise_addr" {
			config.AdvertiseAddr = key.MustString(config.AdvertiseAddr)
		} else if name == "listen_addr" {
			config.ListenAddr = key.MustString(config.ListenAddr)
		} else if name == "log_file" {
			config.LogFile = key.MustString(config.LogFile)
		} else if name == "log_stderr" {
			config.LogStderr = key.MustBool(config.LogStderr)
		} else if name == "http_addr" {
			config.HTTPAddr = key.MustString(config.HTTPAddr)
		} else if name == "log_level" {
			config.LogLevel = key.MustString(config.LogLevel)
		} else {
			gwlog.Fatalf("section %s has unknown key: %s", sec.Name(), key.Name())
		}
	}
	return
}

func readStorageConfig(sec *ini.Section, config *StorageConfig) {
	// setup default values
	config.Type = "mongodb"
	config.DB = _DEFAULT_STORAGE_DB
	config.Url = ""
	config.StartNodes = common.StringSet{}

	for _, key := range sec.Keys() {
		name := strings.ToLower(key.Name())
		if name == "type" {
			config.Type = key.MustString(config.Type)
		} else if name == "url" {
			config.Url = key.MustString(config.Url)
		} else if name == "db" {
			config.DB = key.MustString(config.DB)
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

	validateKVDBConfig(config)
}

func validateKVDBConfig(config *KVDBConfig) {
	if config.Type == "" {
		// KVDB not enabled, it's OK
	} else if config.Type == "mongodb" {
		// must set DB and Collection for mongodb
		if config.Url == "" || config.DB == "" || config.Collection == "" {
			gwlog.Fatalf("invalid %s KVDB config:\n%s", config.Type, DumpPretty(config))
		}
	} else if config.Type == "redis" {
		if config.Url == "" {
			gwlog.Fatalf("invalid %s KVDB config:\n%s", config.Type, DumpPretty(config))
		}
		_, err := strconv.Atoi(config.DB) // make sure db is integer for redis
		if err != nil {
			gwlog.Panic(errors.Wrap(err, "redis db must be integer"))
		}
	} else if config.Type == "redis_cluster" {
		if len(config.StartNodes) == 0 {
			gwlog.Fatalf("must have at least 1 start_nodes for [kvdb].redis_cluster")
		}
		for s := range config.StartNodes {
			if s == "" {
				gwlog.Fatalf("start_nodes must not be empty")
			}
		}
	} else {
		gwlog.Fatalf("unknown storage type: %s", config.Type)
	}
}

func readDebugConfig(sec *ini.Section, config *DebugConfig) {
	config.Debug = false

	for _, key := range sec.Keys() {
		name := strings.ToLower(key.Name())
		if name == "debug" {
			config.Debug = key.MustBool(config.Debug)
		} else {
			gwlog.Fatalf("section %s has unknown key: %s", sec.Name(), key.Name())
		}
	}
}

func checkConfigError(err error, msg string) {
	if err != nil {
		if msg == "" {
			msg = err.Error()
		}
		gwlog.Fatalf("read config error: %s", msg)
	}
}

func validateStorageConfig(config *StorageConfig) {
	if config.Type == "mongodb" {
		if config.Url == "" {
			gwlog.Fatalf("url is not set in %s storage config", config.Type)
		}
		if config.DB == "" {
			gwlog.Fatalf("db is not set in %s storage config", config.Type)
		}
	} else {
		gwlog.Fatalf("unknown storage type: %s", config.Type)
	}
}

func validateConfig(config *GoWorldConfig) {
	deploymentConfig := &config.Deployment
	if deploymentConfig.DesiredGates <= 0 {
		gwlog.Fatalf("[deployment].desired_gates is %d, which must be positive", deploymentConfig.DesiredGates)
	}

	if deploymentConfig.DesiredGames <= 0 {
		gwlog.Fatalf("[deployment].desired_games is %d, which must be positive", deploymentConfig.DesiredGames)
	}

	dispatchersNum := deploymentConfig.DesiredDispatchers
	if dispatchersNum != len(config._Dispatchers) {
		gwlog.Panicf("[deployment].desired_dispatchers is %d, but find %d dispatcher section in config file", dispatchersNum, len(config._Dispatchers))
	}
	if dispatchersNum <= 0 {
		gwlog.Fatalf("dispatcher not found in config file, must has at least 1 dispatcher")
	}

	for dispatcherid := 1; dispatcherid <= dispatchersNum; dispatcherid++ {
		if _, ok := config._Dispatchers[uint16(dispatcherid)]; !ok {
			gwlog.Fatalf("found %d dispatchers in config file, but dispatcher%d is not found. dispatcherid must be 1~%d", dispatchersNum, dispatcherid, dispatchersNum)
		}
	}
}
