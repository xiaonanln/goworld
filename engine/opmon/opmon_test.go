package opmon

import (
	"testing"
	"time"
)

func TestOpMon(t *testing.T) {
	op := StartOperation("test")
	op.Finish(time.Millisecond)
	monitor.Dump()
}
