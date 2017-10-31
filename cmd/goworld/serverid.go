package main

import (
	"path"
	"path/filepath"
	"strings"
)

type ServerID string

func (sid ServerID) Path() string {
	serverPath := strings.Split(string(sid), "/")
	serverPath = append([]string{env.GoWorldRoot}, serverPath...)
	return filepath.Join(serverPath...)
}

func (sid ServerID) Name() string {
	_, file := path.Split(string(sid))
	return file
}
