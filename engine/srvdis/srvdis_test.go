package srvdis

import (
	"context"
	"testing"
	"time"

	"sync"

	"github.com/xiaonanln/goworld/engine/gwlog"
)

type testService struct {
	sync.Mutex
	services [][]string
}

func TestStartup(t *testing.T) {
	ctx := context.Background()
	ctx, _ = context.WithTimeout(ctx, time.Second*5)
	ts := &testService{}
	Startup(ctx, []string{"http://127.0.0.1:2379"}, "/testns", 2, ts)
	Register("testServiceType", "testService", ServiceRegisterInfo{"localhost:12345"})

	<-ctx.Done()
	time.Sleep(time.Second)
	ts.Lock()
	if len(ts.services) == 0 {
		t.Errorf("service discovery failed")
	} else {
		srv := ts.services[0]
		if srv[0] != "testServiceType" || srv[1] != "testService" || srv[2] != "localhost:12345" {
			t.Errorf("service discovered, but wrong")
		}
	}
	ts.Unlock()
}

func (ts *testService) OnServiceDiscovered(srvtype string, srvid string, srvaddr string) {
	ts.Lock()
	ts.services = append(ts.services, []string{srvtype, srvid, srvaddr})
	ts.Unlock()
}

func (ts *testService) OnServiceOutdated(srvtype string, srvid string) {
	gwlog.Infof("testService: OnServiceOutdated: %s.%s", srvtype, srvid)
}
