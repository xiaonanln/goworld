package main

import (
	"fmt"
	"os"
	"sync"
	"time"
)

var (
	profLock             sync.Mutex
	profSectionStartTime time.Time
	thingMaxTime         map[string]time.Duration
	thingMinTime         map[string]time.Duration
	thingTimes           map[string]time.Duration
	thingCounts          map[string]int
)

func recordThingTime(thing string, d time.Duration) {
	profLock.Lock()
	now := time.Now()
	if now.Sub(profSectionStartTime) >= time.Second {
		// start new profile section
		dumpThingProfile()

		profSectionStartTime = now
		thingTimes = map[string]time.Duration{}
		thingCounts = map[string]int{}
		thingMinTime = map[string]time.Duration{}
		thingMaxTime = map[string]time.Duration{}
	}

	thingCounts[thing] += 1
	thingTimes[thing] += d
	if oldMaxTime, ok := thingMaxTime[thing]; !ok || oldMaxTime < d {
		thingMaxTime[thing] = d
	}
	if oldMinTime, ok := thingMaxTime[thing]; !ok || oldMinTime > d {
		thingMinTime[thing] = d
	}

	profLock.Unlock()
}

func dumpThingProfile() {
	for thing, count := range thingCounts {
		totalTime := thingTimes[thing]
		fmt.Fprintf(os.Stdout, "> %-32s *%d AVG %s RANGE %s ~ %s\n", thing, count, totalTime/time.Duration(count), thingMinTime[thing], thingMaxTime[thing])
	}
	fmt.Fprintln(os.Stdout, "===============================================================================")
}
