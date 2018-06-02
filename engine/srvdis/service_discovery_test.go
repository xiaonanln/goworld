package srvdis

import (
	"context"
	"testing"
	"time"

	"github.com/xiaonanln/goworld/engine/gwlog"
)

type testService struct {
}

func TestStartup(t *testing.T) {
	ctx := context.Background()
	ctx, _ = context.WithTimeout(ctx, time.Second*5)
	Startup(ctx, []string{"http://127.0.0.1:2379"}, "/testns", &testService{})
	<-ctx.Done()
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
	gwlog.Infof("testService: OnServiceDiscovered: %s.%s", srvtype, srvid)
}

func (ts *testService) OnServiceOutdated(srvtype string, srvid string) {
	gwlog.Infof("testService: OnServiceOutdated: %s.%s", srvtype, srvid)
}
