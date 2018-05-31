package srvdis

import "testing"

func TestStartup(t *testing.T) {
	Startup([]string{"http://127.0.0.1:2379"}, nil)
}
