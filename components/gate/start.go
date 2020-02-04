package gate

import (
	"flag"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	"github.com/xiaonanln/goworld/engine/config"
	"github.com/xiaonanln/goworld/engine/gwlog"
	"github.com/xiaonanln/goworld/engine/post"
)

var (
	args struct {
		gateid          uint16
		configFile      string
		logLevel        string
		runInDaemonMode bool
		//listenAddr      string
	}
	gateService *GateService
	signalChan  = make(chan os.Signal, 1)
)

func parseArgs() {
	var gateIdArg int
	flag.IntVar(&gateIdArg, "gid", 0, "set gateid")
	flag.StringVar(&args.configFile, "configfile", "", "set config file path")
	flag.StringVar(&args.logLevel, "log", "", "set log level, will override log level in config")
	flag.BoolVar(&args.runInDaemonMode, "d", false, "run in daemon mode")
	//flag.StringVar(&args.listenAddr, "listen-addr", "", "set listen address for gate, overriding listen_addr in config file")
	flag.Parse()
	args.gateid = uint16(gateIdArg)
}

func verifyGateConfig(gateConfig *config.GateConfig) {
}

func setupSignals() {
	gwlog.Infof("Setup signals ...")
	signal.Ignore(syscall.Signal(10), syscall.Signal(12), syscall.SIGPIPE, syscall.SIGHUP)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for {
			sig := <-signalChan
			if sig == syscall.SIGINT || sig == syscall.SIGTERM {
				// terminating gate ...
				gwlog.Infof("Terminating gate service ...")
				post.Post(func() {
					gateService.terminate()
				})

				gateService.terminated.Wait()
				gwlog.Infof("Gate %d terminated gracefully.", args.gateid)
				os.Exit(0)
			} else {
				gwlog.Errorf("unexpected signal: %s", sig)
			}
		}
	}()
}

type gateDispatcherClientDelegate struct {
}
