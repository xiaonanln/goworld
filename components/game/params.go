package game

import "time"

const (
	DISPATCHER_CLIENT_PACKET_QUEUE_SIZE = 0                     // packet queue size
	TICK_INTERVAL                       = time.Millisecond * 10 // game tick interval => affects timer resolution
)
