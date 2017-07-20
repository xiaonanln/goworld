package binutil

import (
	"fmt"
	"net/http"

	"io"
	"os"

	"github.com/xiaonanln/goworld/gwlog"
)

func SetupPprofServer(ip string, port int) {
	if port == 0 {
		// pprof not enabled
		gwlog.Info("pprof server not enabled")
		return
	}

	pprofHost := fmt.Sprintf("%s:%d", ip, port)
	gwlog.Info("pprof server listening on http://%s/debug/pprof/ ... available commands: ", pprofHost)
	gwlog.Info("    go tool pprof http://%s/debug/pprof/heap", pprofHost)
	gwlog.Info("    go tool pprof http://%s/debug/pprof/profile", pprofHost)

	go func() {
		http.ListenAndServe(pprofHost, nil)
	}()
}

func SetupGWLog(logLevel string, logFile string, logStderr bool) {
	gwlog.Info("Set log level to %s", logLevel)
	gwlog.SetLevel(gwlog.StringToLevel(logLevel))

	outputWriters := make([]io.Writer, 0, 2)
	if logFile != "" {
		f, err := os.OpenFile(logFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
		if err != nil {
			panic(err)
		}
		outputWriters = append(outputWriters, f)
	}
	if logStderr {
		outputWriters = append(outputWriters, os.Stderr)
	}

	if len(outputWriters) == 1 {
		gwlog.SetOutput(outputWriters[0])
	} else {
		gwlog.SetOutput(io.MultiWriter(outputWriters...))
	}
}
