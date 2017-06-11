package consts

import "time"

const (
	// For Game & Gate
	DISPATCHER_CLIENT_PACKET_QUEUE_SIZE = 0 // packet queue size
	// For Game
	GAME_TICK_INTERVAL = time.Millisecond * 10 // game tick interval => affects timer resolution
	// For Storage
	//STORAGE_SAVE_QUEUE_SIZE = 100
	//STORAGE_LOAD_QUEUE_SIZE = 1000
)
