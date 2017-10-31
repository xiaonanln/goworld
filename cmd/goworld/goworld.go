package main

import (
	"flag"
	"os"
	"strings"
)

var args struct {
}

func parseArgs() {
	//flag.StringVar(&args.configFile, "configfile", "", "set config file path")

	flag.Parse()
}

func main() {
	parseArgs()
	args := flag.Args()
	showMsg("arguments: %s", strings.Join(args, " "))

	detectGoWorldPath()

	if len(args) == 0 {
		showMsg("no command to execute")
		flag.Usage()
		os.Exit(1)
	}

	cmd := args[0]
	if cmd == "build" {
		if len(args) != 2 {
			showMsgAndQuit("should specify one server id")
		}

		build(ServerID(args[1]))
	} else if cmd == "start" {
		if len(args) != 2 {
			showMsgAndQuit("should specify one server id")
		}

		start(ServerID(args[1]))
	} else if cmd == "stop" {
		if len(args) != 2 {
			showMsgAndQuit("should specify one server id")
		}
		stop(ServerID(args[1]))
	} else if cmd == "reload" {

	} else if cmd == "kill" {

	} else if cmd == "status" {
	} else {
		showMsgAndQuit("unknown command: %s", cmd)
	}
}
