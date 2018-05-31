package srvdis

import (
	"context"
	"testing"
	"time"
)

type testService struct {
}

func TestStartup(t *testing.T) {
	ctx := context.Background()
	Startup(ctx, []string{"http://127.0.0.1:2379"}, "/testns", &testService{})
	time.Sleep(time.Second * 30)
}

func (ts *testService) ServiceType() string {
	return "testServiceType"
}

func (ts *testService) ServiceId() string {
	return "testService"
}
