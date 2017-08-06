package consts

import "time"

// Optimizations
const (
	OPTIMIZE_LOCAL_ENTITIES = false // should be true for performance, set to false for testing
)

// Tunable Options
const (
	// For Packets Send & Recv
	PACKET_PAYLOAD_LEN_COMPRESS_THRESHOLD = 512

	// For Dispatcher
	DISPATCHER_CLIENT_PROXY_WRITE_BUFFER_SIZE = 1024 * 1024
	DISPATCHER_CLIENT_PROXY_READ_BUFFER_SIZE  = 1024 * 1024
	ENTITY_PENDING_PACKET_QUEUE_MAX_LEN       = 1000

	// For Game & Gate
	GAME_SERVICE_PACKET_QUEUE_SIZE = 10000 // packet queue size
	// For Game
	GAME_SERVICE_TICK_INTERVAL = time.Millisecond * 10 // server tick interval => affect timer resolution

	DISPATCHER_CLIENT_WRITE_BUFFER_SIZE = 1024 * 1024
	DISPATCHER_CLIENT_READ_BUFFER_SIZE  = 1024 * 1024

	// For Gate Service
	CLIENT_PROXY_WRITE_BUFFER_SIZE = 1024 * 1024
	CLIENT_PROXY_READ_BUFFER_SIZE  = 1024 * 1024
	COMPRESS_WRITER_POOL_SIZE      = 100

	//SAVE_INTERVAL      = time.Minute * 5 // Save interval of entities

	ENTER_SPACE_REQUEST_TIMEOUT    = DISPATCHER_MIGRATE_TIMEOUT + time.Minute // enter space should finish in limited seconds
	DISPATCHER_MIGRATE_TIMEOUT     = time.Minute * 5
	DISPATCHER_LOAD_TIMEOUT        = time.Minute * 5
	DISPATCHER_FREEZE_GAME_TIMEOUT = time.Minute * 5
	// For Storage
	// For Operation Monitor
	OPMON_DUMP_INTERVAL = time.Second * 10
)

// Debug Options
const (
	DEBUG_PACKETS      = false
	DEBUG_SPACES       = false
	DEBUG_SAVE_LOAD    = false
	DEBUG_CLIENTS      = false
	DEBUG_MIGRATE      = false
	DEBUG_PACKET_ALLOC = false
	DEBUG_FILTER_PROP  = false
)

//  System level configurations
const (
	DEBUG_MODE = false
)
