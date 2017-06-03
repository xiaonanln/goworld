package goworld

import (
	"time"
	"os"
)

func Run() {
	for {
		os.Stdout.Write([]byte("."))
		time.Sleep(time.Second)
	}
}