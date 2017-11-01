package binutil

import (
	"fmt"
	"net/http"

	"io"
	"os"

	"github.com/natefinch/lumberjack"
	"github.com/xiaonanln/goworld/engine/gwlog"
	"golang.org/x/net/websocket"
)

// SetupHTTPServer starts the HTTP server for go tool pprof and websockets
func SetupHTTPServer(ip string, port int, wsHandler func(ws *websocket.Conn)) {
	setupHTTPServer(ip, port, wsHandler, "", "")
}

// SetupHTTPServerTLS starts the HTTPs server for go tool pprof and websockets
func SetupHTTPServerTLS(ip string, port int, wsHandler func(ws *websocket.Conn), certFile string, keyFile string) {
	setupHTTPServer(ip, port, wsHandler, certFile, keyFile)
}

func setupHTTPServer(ip string, port int, wsHandler func(ws *websocket.Conn), certFile string, keyFile string) {
	if port == 0 {
		// pprof not enabled
		gwlog.Infof("pprof server not enabled")
		return
	}

	httpHost := fmt.Sprintf("%s:%d", ip, port)
	gwlog.Infof("http server listening on %s", httpHost)
	gwlog.Infof("pprof http://%s/debug/pprof/ ... available commands: ", httpHost)
	gwlog.Infof("    go tool pprof http://%s/debug/pprof/heap", httpHost)
	gwlog.Infof("    go tool pprof http://%s/debug/pprof/profile", httpHost)
	if keyFile != "" || certFile != "" {
		gwlog.Infof("TLS is enabled on http: key=%s, cert=%s", keyFile, certFile)
	}

	//http.Handle("/", http.FileServer(http.Dir(".")))
	if wsHandler != nil {
		http.Handle("/ws", websocket.Handler(wsHandler))
	}

	go func() {
		if keyFile == "" && certFile == "" {
			http.ListenAndServe(httpHost, nil)
		} else {
			http.ListenAndServeTLS(httpHost, certFile, keyFile, nil)
		}
	}()
}

// SetupGWLog setup the GoWord log system
func SetupGWLog(component string, logLevel string, logFile string, logStderr bool) {
	gwlog.SetSource(component)
	gwlog.Infof("Set log level to %s", logLevel)
	gwlog.SetLevel(gwlog.StringToLevel(logLevel))

	outputWriters := make([]io.Writer, 0, 2)
	if logFile != "" {
		var logFileWriter io.Writer
		logFileWriter = &lumberjack.Logger{
			Filename:   logFile,
			MaxSize:    100, // megabytes
			MaxBackups: 100,
			MaxAge:     30, //days
			Compress:   true,
		}

		logFileWriter.(*lumberjack.Logger).Rotate() // rotate immediately
		outputWriters = append(outputWriters, logFileWriter)
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
