package main

import (
	"path"
	"path/filepath"
	"strings"
)

// ServerID represents a server.
// It's the 2nd argument of `goworld.exe` CLI.
type ServerID string

// Path returns the path to the server
func (sid ServerID) Path() string {
	serverPath := strings.Split(string(sid), "/")
	serverPath = append([]string{env.GoWorldRoot}, serverPath...)
	return filepath.Join(serverPath...)
}

// Name returns the name of the server
func (sid ServerID) Name() string {
	_, file := path.Split(string(sid))
	return file
}
