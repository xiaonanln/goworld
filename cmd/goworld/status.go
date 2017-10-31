package main

import "github.com/shirou/gopsutil/process"

func detectRunningServer() {
	pids, err := process.Pids()
	checkErrorOrQuit(err, "get pids failed")
	for _, pid := range pids {
		proc, err := process.NewProcess(pid)
		checkErrorOrQuit(err, "new process failed")
		exe, _ := proc.Exe()
		println("process", exe)
	}
}
