// +build windows

package main

import (
	_ "github.com/go-ole/go-ole" // so that dep can resolve versions correctly
)

const (
	IsWindows    = true
	ExecutiveExt = ".exe"
)
