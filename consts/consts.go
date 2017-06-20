package consts

import "time"

// Tunable Options
const (
	// For Server & Gate
	DISPATCHER_CLIENT_PACKET_QUEUE_SIZE = 100 // packet queue size
	// For Server
	SERVER_TICK_INTERVAL = time.Millisecond * 10 // server tick interval => affects timer resolution
	SAVE_INTERVAL        = time.Minute * 1       // TODO: config save interval by goworld.ini
	//SAVE_INTERVAL      = time.Minute * 5 // Save interval of entities
	ENTER_SPACE_REQUEST_TIMEOUT = time.Second * 10 // enter space should finish in limited seconds
	// For Storage
)

// Debug Options
const (
	DEBUG_PACKETS   = false
	DEBUG_SPACES    = true
	DEBUG_SAVE_LOAD = false
	DEBUG_CLIENTS   = true
	DEBUG_MIGRATE   = true
)
