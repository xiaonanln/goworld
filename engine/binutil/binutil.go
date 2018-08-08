package binutil

import (
	"net/http"
	"syscall"

	"github.com/xiaonanln/goworld/engine/gwlog"
	"golang.org/x/net/websocket"
)

const (
	// FreezeSignal syscall used to freeze server
	FreezeSignal = syscall.SIGHUP
)

// SetupHTTPServer starts the HTTP server for go tool pprof and websockets
func SetupHTTPServer(listenAddr string, wsHandler func(ws *websocket.Conn)) {
	setupHTTPServer(listenAddr, wsHandler, "", "")
}

// SetupHTTPServerTLS starts the HTTPs server for go tool pprof and websockets
func SetupHTTPServerTLS(listenAddr string, wsHandler func(ws *websocket.Conn), certFile string, keyFile string) {
	setupHTTPServer(listenAddr, wsHandler, certFile, keyFile)
}

func setupHTTPServer(listenAddr string, wsHandler func(ws *websocket.Conn), certFile string, keyFile string) {
	gwlog.Infof("http server listening on %s", listenAddr)
	gwlog.Infof("pprof http://%s/debug/pprof/ ... available commands: ", listenAddr)
	gwlog.Infof("    go tool pprof http://%s/debug/pprof/heap", listenAddr)
	gwlog.Infof("    go tool pprof http://%s/debug/pprof/profile", listenAddr)
	if keyFile != "" || certFile != "" {
		gwlog.Infof("TLS is enabled on http: key=%s, cert=%s", keyFile, certFile)
	}

	//http.Handle("/", http.FileServer(http.Dir(".")))
	if wsHandler != nil {
		http.Handle("/ws", websocket.Handler(wsHandler))
	}

	go func() {
		if keyFile == "" && certFile == "" {
			http.ListenAndServe(listenAddr, nil)
		} else {
			http.ListenAndServeTLS(listenAddr, certFile, keyFile, nil)
		}
	}()
}

// SetupGWLog setup the GoWord log system
func SetupGWLog(component string, logLevel string, logFile string, logStderr bool) {
	gwlog.SetSource(component)
	gwlog.Infof("Set log level to %s", logLevel)
	gwlog.SetLevel(gwlog.ParseLevel(logLevel))

	var outputs []string
	if logStderr {
		outputs = append(outputs, "stderr")
	}
	if logFile != "" {
		outputs = append(outputs, logFile)
	}
	gwlog.SetOutput(outputs)

	//outputWriters := make([]io.Writer, 0, 2)
	//if logFile != "" {
	//	fileWriter, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	//	if err != nil {
	//		gwlog.Fatalf("open log file %s failed: %v", logFile, err)
	//	}
	//	outputWriters = append(outputWriters, fileWriter)
	//}
	//
	//if logStderr {
	//	outputWriters = append(outputWriters, os.Stderr)
	//}
	//
	//if len(outputWriters) == 1 {
	//	gwlog.SetOutput(outputWriters[0])
	//} else {
	//	gwlog.SetOutput(io.MultiWriter(outputWriters...))
	//}
}

func PrintSupervisorTag(tag string) {
	curlvl := gwlog.GetLevel()
	if curlvl != gwlog.DebugLevel && curlvl != gwlog.InfoLevel {
		gwlog.SetLevel(gwlog.InfoLevel)
	}
	gwlog.Infof("%s", tag)
	if curlvl != gwlog.DebugLevel && curlvl != gwlog.InfoLevel {
		gwlog.SetLevel(curlvl)
	}
}
