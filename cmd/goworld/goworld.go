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
		for _, arg := range args[1:] {
			build(arg)
		}
	}
}
