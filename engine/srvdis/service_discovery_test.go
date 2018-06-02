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
	ctx, _ = context.WithTimeout(ctx, time.Second*3)
	ts := &testService{}
	Startup(ctx, []string{"http://127.0.0.1:2379"}, "/testns", ts)
	<-ctx.Done()
	ts.Lock()
	if len(ts.services) == 0 {
		gwlog.Errorf("service discovery failed")
	} else {
		srv := ts.services[0]
		if srv[0] != "testServiceType" || srv[1] != "testService" || srv[2] != "localhost:12345" {
			gwlog.Errorf("service discovered, but wrong")
		}
	}
	ts.Unlock()
}

func (ts *testService) ServiceType() string {
	return "testServiceType"
}

func (ts *testService) ServiceId() string {
	return "testService"
}

func (ts *testService) ServiceAddr() string {
	return "localhost:12345"
}

func (ts *testService) ServiceLeaseTTL() int64 {
	return 2
}

func (ts *testService) OnServiceDiscovered(srvtype string, srvid string, srvaddr string) {
	ts.Lock()
	ts.services = append(ts.services, []string{srvtype, srvid, srvaddr})
	ts.Unlock()
}

func (ts *testService) OnServiceOutdated(srvtype string, srvid string) {
	gwlog.Infof("testService: OnServiceOutdated: %s.%s", srvtype, srvid)
}
