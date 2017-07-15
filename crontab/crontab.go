package crontab

import (
	"time"

	"github.com/xiaonanln/goTimer"
	"github.com/xiaonanln/goworld/gwlog"
	"github.com/xiaonanln/goworld/gwutils"
)

const (
	_CRONTAB_TIME_OFFSET = time.Second * 2
)

var (
	cancelledHandles = []Handle{}
	entries          = map[Handle]*entry{}
	nextHandle       = Handle(1)
)

type Handle int // Return value of Register, can be used to cancel the register

type entry struct {
	minute, hour, day, month, dayofweek int
	cb                                  func()
}

func (entry *entry) match(minute int, hour int, day int, month time.Month, weekday time.Weekday) bool {
	if entry.minute >= 0 && entry.minute != minute {
		return false
	}
	if entry.hour >= 0 && entry.hour != hour {
		return false
	}
	if entry.day >= 0 && entry.day != day {
		return false
	}
	if entry.month >= 0 && entry.month != int(month) {
		return false
	}
	if entry.dayofweek >= 0 {
		if entry.dayofweek >= 1 && entry.dayofweek <= 6 {
			if entry.dayofweek != int(weekday) {
				return false
			}
		} else if entry.dayofweek == 0 || entry.dayofweek == 7 {
			if weekday != time.Sunday {
				return false
			}
		} else { // invalid dayofweek, never happen
			return false
		}
	}
	return true
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
	validateTime(minute, hour, day, month, dayofweek)

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

func validateTime(minute, hour, day, month, dayofweek int) {
	if (minute < 0 || minute > 59) && minute != -1 {
		gwlog.Panicf("invalid minute = %d", minute)
	}

	if (hour < 0 || hour > 23) && hour != -1 {
		gwlog.Panicf("invalid hour = %d", hour)
	}
	if (day < 1 || day > 31) && day != -1 {
		gwlog.Panicf("invalid day = %d", day)
	}
	if (month < 1 && month > 12) && month != -1 {
		gwlog.Panicf("invalid month = %d", month)
	}
	if (dayofweek < 0 && dayofweek > 7) && dayofweek != -1 {
		gwlog.Panicf("invalid dayofweek = %d", dayofweek)
	}
}

// Unregister a registered crontab handle
func (h Handle) Unregister() {
	cancelledHandles = append(cancelledHandles, h)
}

func unregisterCancelledHandles() {
	for _, h := range cancelledHandles {
		gwlog.Debug("unregisterCancelledHandles: cancelling %d", h)
		delete(entries, h)
	}
	cancelledHandles = nil
}

// Initialize crontab module, called by engine
func Initialize() {
	now := time.Now()
	sec := now.Second()
	var d time.Duration
	if time.Second*time.Duration(sec) < _CRONTAB_TIME_OFFSET {
		d = _CRONTAB_TIME_OFFSET - time.Second*time.Duration(sec)
	} else {
		d = time.Second*time.Duration(60-sec) + _CRONTAB_TIME_OFFSET
	}

	d -= time.Nanosecond * time.Duration(now.Nanosecond())
	gwlog.Info("current time is %s, will setup repeat time after %s", now, d)
	timer.AddCallback(d, func() {
		setupRepeatTimer()
		check()
	})
}

func check() {
	unregisterCancelledHandles()

	now := time.Now()
	gwlog.Info("%s: crontab: checking %d callbacks ...", now, len(entries))
	dayofweek, month, day, hour, minute := now.Weekday(), now.Month(), now.Day(), now.Hour(), now.Minute()

	for _, entry := range entries {
		if entry.match(minute, hour, day, month, dayofweek) {
			gwutils.RunPanicless(entry.cb)
		}
	}

	unregisterCancelledHandles()
}

func setupRepeatTimer() {
	gwlog.Info("crontab: setup repeat timer at time %s", time.Now())
	timer.AddTimer(time.Minute, check)
}

func genNextHandle() (h Handle) {
	h, nextHandle = nextHandle, nextHandle+1
	return
}
