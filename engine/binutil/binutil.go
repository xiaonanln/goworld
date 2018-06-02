package binutil

import (
	"fmt"
	"net/http"

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
