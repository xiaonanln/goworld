package post

import (
	"sync"

	//"github.com/xiaonanln/goworld/gwlog"
	"github.com/xiaonanln/goworld/engine/gwutils"
)

// PostCallback is the type of functions to be posted
type PostCallback func()

var (
	callbacks []PostCallback
	lock      sync.Mutex
)

// Post a callback which will be executed when other things are done in the main game routine
//
// Post might be called from other goroutine, so we use a lock to protect the data
func Post(f PostCallback) {
	lock.Lock()
	callbacks = append(callbacks, f)
	lock.Unlock()
}

// Tick is called by the main game routine to run all posted functions
func Tick() {
	for { // loop until there is no callbacks posted anymore
		lock.Lock() // lock to check number of callbacks
		if len(callbacks) == 0 {
			lock.Unlock()
			break // all callbacked executed, quit
		}
		// switch callbacks in locked section
		callbacksCopy := callbacks
		callbacks = make([]PostCallback, 0, len(callbacks))
		lock.Unlock()

		for _, f := range callbacksCopy {
			gwutils.RunPanicless(f)
		}
	}
}
