package gwutils

import (
	"fmt"
	"testing"
)

func TestRunPanicless(t *testing.T) {
	RunPanicless(func() {
		panic(1)
	})
	RunPanicless(func() {
		panic(fmt.Errorf("bad"))
	})
}
