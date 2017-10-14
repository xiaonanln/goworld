package async

import (
	"sync"
	"testing"

	"time"

	"github.com/xiaonanln/goworld/engine/post"
)

func TestNewAsyncJob(t *testing.T) {
	var wait sync.WaitGroup
	wait.Add(2)
	AppendAsyncJob("1", func() (res interface{}, err error) {
		wait.Done()
		return 1, nil
	}, func(res interface{}, err error) {
		println("returns", res.(int), err)
		wait.Done()
	})
	wait.Wait()
}

func init() {
	go func() {
		for {
			post.Tick()
			time.Sleep(time.Millisecond)
		}
	}()
}
