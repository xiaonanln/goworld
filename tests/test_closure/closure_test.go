package test_closure

import (
	"fmt"
	"testing"
	"time"

	"github.com/xiaonanln/goTimer"
)

func TestClosure(t *testing.T) {
	a := []int{}

	for i := 0; i < 10; i++ {
		a = append(a, i)
		i, a := i, a
		timer.AddCallback(time.Second+time.Duration(i)*time.Millisecond, func() {
			fmt.Println(i, a)
		})
	}
	timer.StartTicks(time.Millisecond)
	time.Sleep(time.Second * 2)
}
