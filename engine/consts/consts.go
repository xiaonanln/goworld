package consts

import "time"

// Optimizations
const (
	OPTIMIZE_LOCAL_ENTITY_CALL = true // should be true for performance, set to false for testing only
)

// Tunable Options
const (
	// For Underlying Networking
	// BUFFERED_READ_BUFFSIZE is the read buffer size for BufferedReadConnection
	BUFFERED_READ_BUFFSIZE = 16384
	// BUFFERED_WRITE_BUFFSIZE is the write buffer size for BufferedWriteConnection
	BUFFERED_WRITE_BUFFSIZE = 16384

	// For Dispatcher
	// DISPATCHER_GC_PERCENT is the GC percent for dispatcher
	DISPATCHER_GC_PERCENT = 1000
	// DISPATCHER_CLIENT_PROXY_WRITE_BUFFER_SIZE is dispatcher client proxies' write buffer size
	DISPATCHER_CLIENT_PROXY_WRITE_BUFFER_SIZE = 1024 * 1024
	// DISPATCHER_CLIENT_PROXY_READ_BUFFER_SIZE is dispatcher client proxies' read buffer size
	DISPATCHER_CLIENT_PROXY_READ_BUFFER_SIZE = 1024 * 1024
	// GAME_PENDING_PACKET_QUEUE_MAX_LEN is the maxium number of packets in pending queue when game is blocked
	GAME_PENDING_PACKET_QUEUE_MAX_LEN = 1000000
	// ENTITY_PENDING_PACKET_QUEUE_MAX_LEN is the maxium number of packets in pending queue when entity is blocked
	ENTITY_PENDING_PACKET_QUEUE_MAX_LEN = 1000

	DISPATCHER_SERVICE_PACKET_QUEUE_SIZE = 10000
	// DISPATCHER_SERVICE_TICK_INTERVAL is the tick interval for dispatcher service's main routine.
	DISPATCHER_SERVICE_TICK_INTERVAL = time.Millisecond * 5 // server tick interval => affect timer resolution

	// For Game Service
	// GAME_SERVICE_PACKET_QUEUE_SIZE is the max packet queue length for game service
	GAME_SERVICE_PACKET_QUEUE_SIZE = 10000 // packet queue size
	// GAME_SERVICE_TICK_INTERVAL is the tick interval to tick timers in game service
	GAME_SERVICE_TICK_INTERVAL = time.Millisecond * 5 // server tick interval => affect timer resolution

	// DISPATCHER_CLIENT_WRITE_BUFFER_SIZE is the writer buffer size for gates/games' connections to dispatcher
	DISPATCHER_CLIENT_WRITE_BUFFER_SIZE = 1024 * 1024
	// DISPATCHER_CLIENT_READ_BUFFER_SIZE is the read buffer size for gates/games' connections to dispatcher
	DISPATCHER_CLIENT_READ_BUFFER_SIZE = 1024 * 1024

	// For Gate Service
	// GATE_SERVICE_PACKET_QUEUE_SIZE is the packet queue size of gate service
	GATE_SERVICE_PACKET_QUEUE_SIZE = 10000
	// GATE_SERVICE_TICK_INTERVAL is the tick interval to tick timers in gate service
	GATE_SERVICE_TICK_INTERVAL = time.Millisecond * 5 // server tick interval => affect timer resolution
	// CLIENT_PROXY_WRITE_BUFFER_SIZE is the write buffer size for gates' client proxies
	CLIENT_PROXY_WRITE_BUFFER_SIZE = 1024 * 1024
	// CLIENT_PROXY_READ_BUFFER_SIZE is the read buffer size for gates' client proxies
	CLIENT_PROXY_READ_BUFFER_SIZE = 1024 * 1024
	// CLIENT_PROXY_SET_TCP_NO_DELAY = true sets client proxies to TcpNoDelay
	CLIENT_PROXY_SET_TCP_NO_DELAY = true

	// ENTER_SPACE_REQUEST_TIMEOUT is the timeout for enter space request
	ENTER_SPACE_REQUEST_TIMEOUT = DISPATCHER_MIGRATE_TIMEOUT + time.Minute // enter space should finish in limited seconds
	// DISPATCHER_MIGRATE_TIMEOUT is timeout for entity migration
	DISPATCHER_MIGRATE_TIMEOUT = time.Minute
	// DISPATCHER_LOAD_TIMEOUT is timeout for loading entity
	DISPATCHER_LOAD_TIMEOUT = time.Minute
	// DISPATCHER_FREEZE_GAME_TIMEOUT is timeout for freezing & restoring game
	DISPATCHER_FREEZE_GAME_TIMEOUT = time.Second * 10
	// For Storage
	// For Operation Monitor
	// OPMON_DUMP_INTERVAL is the interval to print opmon infos to output
	OPMON_DUMP_INTERVAL = 0
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
	// DEBUG_FILTER_PROP prints filter props debug logs
	DEBUG_FILTER_PROP = false
)

//  System level configurations
const (
	// DEBUG_MODE = true turns on debug mode
	DEBUG_MODE = true
)

// Async configurations
const (
	ASYNC_JOB_QUEUE_MAXLEN = 10000
)

// KCP Options
const (
	KCP_NO_DELAY                       = 1  // Whether nodelay mode is enabled, 0 is not enabled; 1 enabled
	KCP_INTERNAL_UPDATE_TIMER_INTERVAL = 10 // Protocol internal work interval, in milliseconds, such as 10 ms or 20 ms.
	KCP_ENABLE_FAST_RESEND             = 2  // Fast retransmission mode, 0 represents off by default, 2 can be set (2 ACK spans will result in direct retransmission)
	KCP_DISABLE_CONGESTION_CONTROL     = 1  // Whether to turn off flow control, 0 represents “Do not turn off” by default, 1 represents “Turn off”.

	KCP_SET_STREAM_MODE  = true
	KCP_SET_WRITE_DELAY  = true
	KCP_SET_ACK_NO_DELAY = true
)

const (
	DISPATCHER_STARTED_TAG = "<!--XSUPERVISOR:BEGIN--> DISPATCHER STARTED <!--XSUPERVISOR:END-->"
	GAME_STARTED_TAG       = "<!--XSUPERVISOR:BEGIN--> GAME STARTED <!--XSUPERVISOR:END-->"
	GATE_STARTED_TAG       = "<!--XSUPERVISOR:BEGIN--> GATE STARTED <!--XSUPERVISOR:END-->"
)
