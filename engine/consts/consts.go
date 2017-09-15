package consts

import "time"

// Optimizations
const (
	OPTIMIZE_LOCAL_ENTITIES = false // should be true for performance, set to false for testing
)

// Tunable Options
const (
	// For Underlying Networking
	// BUFFERED_READ_BUFFSIZE is the read buffer size for BufferedReadConnection
	BUFFERED_READ_BUFFSIZE = 16384
	// BUFFERED_WRITE_BUFFSIZE is the write buffer size for BufferedWriteConnection
	BUFFERED_WRITE_BUFFSIZE = 16384

	// For Packets Send & Recv
	// PACKET_PAYLOAD_LEN_COMPRESS_THRESHOLD is the minimal packet payload length that should be compressed
	PACKET_PAYLOAD_LEN_COMPRESS_THRESHOLD = 512

	// For Dispatcher
	// DISPATCHER_CLIENT_PROXY_WRITE_BUFFER_SIZE is dispatcher client proxies' write buffer size
	DISPATCHER_CLIENT_PROXY_WRITE_BUFFER_SIZE = 1024 * 1024
	// DISPATCHER_CLIENT_PROXY_READ_BUFFER_SIZE is dispatcher client proxies' read buffer size
	DISPATCHER_CLIENT_PROXY_READ_BUFFER_SIZE = 1024 * 1024
	// ENTITY_PENDING_PACKET_QUEUE_MAX_LEN is the maxium number of packets in pending queue when entity is blocked
	ENTITY_PENDING_PACKET_QUEUE_MAX_LEN = 1000
	// MAX_ENTITY_SYNC_INFOS_CACHE_SIZE_PER_GAME is maxium number of bytes of entity sync info cached for each game
	MAX_ENTITY_SYNC_INFOS_CACHE_SIZE_PER_GAME = 1024 * 1024

	// For Game Service
	// GAME_SERVICE_PACKET_QUEUE_SIZE is the max packet queue length for game service
	GAME_SERVICE_PACKET_QUEUE_SIZE = 10000 // packet queue size
	// GAME_SERVICE_TICK_INTERVAL is the tick interval to tick timers in game service
	GAME_SERVICE_TICK_INTERVAL = time.Millisecond * 10 // server tick interval => affect timer resolution

	// DISPATCHER_CLIENT_WRITE_BUFFER_SIZE is the writer buffer size for gates/games' connections to dispatcher
	DISPATCHER_CLIENT_WRITE_BUFFER_SIZE = 1024 * 1024
	// DISPATCHER_CLIENT_READ_BUFFER_SIZE is the read buffer size for gates/games' connections to dispatcher
	DISPATCHER_CLIENT_READ_BUFFER_SIZE = 1024 * 1024

	// For Gate Service
	// CLIENT_PROXY_WRITE_BUFFER_SIZE is the write buffer size for gates' client proxies
	CLIENT_PROXY_WRITE_BUFFER_SIZE = 1024 * 1024
	// CLIENT_PROXY_READ_BUFFER_SIZE is the read buffer size for gates' client proxies
	CLIENT_PROXY_READ_BUFFER_SIZE = 1024 * 1024
	// COMPRESS_WRITER_POOL_SIZE is number of write compressors in the pool for gate
	COMPRESS_WRITER_POOL_SIZE = 100
	// CLIENT_PROXY_SET_TCP_NO_DELAY = true sets client proxies to TcpNoDelay
	CLIENT_PROXY_SET_TCP_NO_DELAY = true

	//SAVE_INTERVAL      = time.Minute * 5 // Save interval of entities

	// ENTER_SPACE_REQUEST_TIMEOUT is the timeout for enter space request
	ENTER_SPACE_REQUEST_TIMEOUT = DISPATCHER_MIGRATE_TIMEOUT + time.Minute // enter space should finish in limited seconds
	// DISPATCHER_MIGRATE_TIMEOUT is timeout for entity migration
	DISPATCHER_MIGRATE_TIMEOUT = time.Minute * 5
	// DISPATCHER_LOAD_TIMEOUT is timeout for loading entity
	DISPATCHER_LOAD_TIMEOUT = time.Minute * 5
	// DISPATCHER_FREEZE_GAME_TIMEOUT is timeout for freezing & restoring game
	DISPATCHER_FREEZE_GAME_TIMEOUT = time.Minute * 5
	// For Storage
	// For Operation Monitor
	// OPMON_DUMP_INTERVAL is the interval to print opmon infos to output
	OPMON_DUMP_INTERVAL = 0

	// For Snappy Compress
	SNAPPY_CHECKSUM_ENABLED   = false
	MIN_DATA_SIZE_TO_COMPRESS = 1024
)

// Debug Options
const (
	// DEBUG_PACKETS prints packet send/recv debug logs
	DEBUG_PACKETS = false
	// DEBUG_SPACES prints space operation debug logs
	DEBUG_SPACES = false
	// DEBUG_SAVE_LOAD prints save & load debug logs
	DEBUG_SAVE_LOAD = false
	// DEBUG_CLIENTS prints clients operation debug logs
	DEBUG_CLIENTS = false
	// DEBUG_MIGRATE prints migration debug logs
	DEBUG_MIGRATE = false
	// DEBUG_PACKET_ALLOC prints  packet allocation debug logs
	DEBUG_PACKET_ALLOC = false
	// DEBUG_FILTER_PROP prints filter props debug logs
	DEBUG_FILTER_PROP = false
)

//  System level configurations
const (
	// DEBUG_MODE = true turns on debug mode
	DEBUG_MODE = false
)
