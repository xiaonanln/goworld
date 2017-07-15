package crontab

import "time"

const (
	_CRONTAB_TIME_OFFSET = time.Second * 2
)

var (
	entries    map[Handle]*entry
	nextHandle = Handle(1)
)

type Handle int

type entry struct {
	minute, hour, day, month, dayofweek int
	cb                                  func()
}

// Register a callack which will be executed when time condition is satisfied
//
// param minute: time condition satisfied on the specified minute, or every minute if minute == -1
// param hour: time condition satisfied on the specified hour, or any hour when hour == -1
// param day: time condition satisfied on the specified day, or any day when day == -1
// param month: time condition satisfied on the specified month, or any month when month == -1
// param dayofweek: time condition satisfied on the specified week day, or any day when dayofweek == -1
// param cb: callback function to be executed when time is satisfied
func Register(minute, hour, day, month, dayofweek int, cb func()) Handle {
	h := genNextHandle()
	entries[h] = &entry{
		minute:    minute,
		hour:      hour,
		day:       day,
		month:     month,
		dayofweek: dayofweek,
		cb:        cb,
	}
	return h
}

func Unregister(h Handle) {
	delete(entries, h)
}

// Initialize crontab module, called by engine
func Initialize() {

}

func genNextHandle() (h Handle) {
	h, nextHandle = nextHandle, nextHandle+1
	return
}
