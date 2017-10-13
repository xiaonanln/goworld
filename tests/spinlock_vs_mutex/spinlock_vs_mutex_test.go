package spinlock_vs_mutex

import (
	"math/rand"
	"sync"
	"testing"
	"time"

	"runtime/pprof"

	"os"

	"github.com/xiaonanln/go-xnsyncutil/xnsyncutil"
	"github.com/xiaonanln/goworld/engine/gwlog"
)

const (
	GOROUTINE_NUM        = 200000
	SLEEP_INTERVAL_MIN   = 500
	SLEEP_INTERVAL_MAX   = 1000
	DUMMY_OP_COUNT       = 10000
	ENABLE_PPROF_PROFILE = false
)

func TestMutex(t *testing.T) {
	if os.Getenv("TRAVIS") != "" {
		t.Skip()
	}

	if ENABLE_PPROF_PROFILE {
		out, err := os.OpenFile("TestMutex.pprof", os.O_CREATE|os.O_TRUNC, 0777)
		if err != nil {
			gwlog.Panic(err)
		}
		defer out.Close()

		pprof.StartCPUProfile(out)
	}

	var lock sync.Mutex
	var wait sync.WaitGroup
	wait.Add(GOROUTINE_NUM)
	for i := 0; i < GOROUTINE_NUM; i++ {
		go func() {
			for i := 0; i < 10; i++ {
				time.Sleep(time.Millisecond * time.Duration(SLEEP_INTERVAL_MIN+rand.Intn((SLEEP_INTERVAL_MAX-SLEEP_INTERVAL_MIN))))
				lock.Lock()
				var dummy int
				for i := 0; i < DUMMY_OP_COUNT; i++ {
					dummy += 1
				}
				lock.Unlock()
			}
			wait.Done()
		}()
	}
	gwlog.Infof("All goroutins are created.")
	wait.Wait()
	if ENABLE_PPROF_PROFILE {
		pprof.StopCPUProfile()
	}
}
func TestRWMutex(t *testing.T) {
	if os.Getenv("TRAVIS") != "" {
		t.Skip()
	}

	if ENABLE_PPROF_PROFILE {
		out, err := os.OpenFile("TestMutex.pprof", os.O_CREATE|os.O_TRUNC, 0777)
		if err != nil {
			gwlog.Panic(err)
		}
		defer out.Close()

		pprof.StartCPUProfile(out)
	}

	var lock sync.RWMutex
	var wait sync.WaitGroup
	wait.Add(GOROUTINE_NUM)
	for i := 0; i < GOROUTINE_NUM; i++ {
		go func() {
			for i := 0; i < 10; i++ {
				time.Sleep(time.Millisecond * time.Duration(SLEEP_INTERVAL_MIN+rand.Intn((SLEEP_INTERVAL_MAX-SLEEP_INTERVAL_MIN))))
				lock.RLock()
				var dummy int
				for i := 0; i < DUMMY_OP_COUNT; i++ {
					dummy += 1
				}
				lock.RUnlock()
			}
			wait.Done()
		}()
	}
	gwlog.Infof("All goroutins are created.")
	wait.Wait()
	if ENABLE_PPROF_PROFILE {
		pprof.StopCPUProfile()
	}
}

func TestSpinLock(t *testing.T) {
	if os.Getenv("TRAVIS") != "" {
		t.Skip()
	}

	if ENABLE_PPROF_PROFILE {
		out, err := os.OpenFile("TestSpinLock.pprof", os.O_CREATE|os.O_TRUNC, 0777)
		if err != nil {
			gwlog.Panic(err)
		}
		defer out.Close()

		pprof.StartCPUProfile(out)
	}

	var lock xnsyncutil.SpinLock
	var wait sync.WaitGroup
	wait.Add(GOROUTINE_NUM)
	for i := 0; i < GOROUTINE_NUM; i++ {
		go func() {
			for i := 0; i < 10; i++ {
				time.Sleep(time.Millisecond * time.Duration(SLEEP_INTERVAL_MIN+rand.Intn((SLEEP_INTERVAL_MAX-SLEEP_INTERVAL_MIN))))
				lock.Lock()
				var dummy int
				for i := 0; i < DUMMY_OP_COUNT; i++ {
					dummy += 1
				}
				lock.Unlock()
			}
			wait.Done()
		}()
	}
	gwlog.Infof("All goroutins are created.")
	wait.Wait()
	if ENABLE_PPROF_PROFILE {
		pprof.StopCPUProfile()
	}
}
