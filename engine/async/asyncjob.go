package async

import (
	"github.com/xiaonanln/goworld/engine/netutil"
)

var (
	jobQueues = map[string]chan func() (interface{}, error){}
)

func init() {
	netutil.ServeForever(func() {
		for {
			_ = <-jobQueue

		}
	})
}

func NewAsyncJob(group string, job func() (interface{}, error)) {

}
