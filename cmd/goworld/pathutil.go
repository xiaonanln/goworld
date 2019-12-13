package main

import (
	"io"
	"os"
	"path/filepath"
)

func isfile(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}

		panic(err)
	}

	return !fi.IsDir()
}

func isdir(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}

		panic(err)
	}

	return fi.IsDir()
}

func isexists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}

		panic(err)
	}
	return true
}

func binPath() string {
	return filepath.Join(env.WorkspaceRoot, "bin")
}

func copyFile(src, dest string) (err error) {
	msg := "Failed to copy default config file."
	source, err := os.Open(src)
	checkErrorOrQuit(err, msg)
	defer func() {
		err = source.Close()
		checkErrorOrQuit(err, msg)
	}()

	destination, err := os.Create(dest)
	checkErrorOrQuit(err, msg)
	defer func() {
		err = destination.Close()
		checkErrorOrQuit(err, msg)
	}()
	_, err = io.Copy(destination, source)
	return
}
