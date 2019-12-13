package main

import (
	"path/filepath"
	"strings"
)

// ServerID represents a server.
// It's the 2nd argument of `goworld` CLI.
type ServerID string

// Path returns the path to the server
func (sid ServerID) Path() string {
	server := strings.Split(string(sid), "/")

	// We first follow Go's conventional workspace directory structure.
	// Where all source lives in the `src` directory.
	parts := append([]string{env.GetSourceDir()}, server...)
	srcDir := filepath.Join(parts...)
	if isdir(srcDir) {
		return srcDir
	}

	// If the source package cannot be found by conventional structure,
	// we then assume that it's using a structure where source are
	// placed inside `goworld` directory.
	parts = append([]string{env.GoWorldRoot}, server...)
	return filepath.Join(parts...)
}

// Name returns the name of the server
func (sid ServerID) Name() string {
	return filepath.Base(string(sid))
}

// BinaryName returns the file name of the server executable
func (sid ServerID) BinaryName() string {
	return sid.Name() + BinaryExtension
}

// BinaryPathName returns the full path of the server executable
func (sid ServerID) BinaryPathName() string {
	return filepath.Join(sid.Path(), sid.BinaryName())
}
