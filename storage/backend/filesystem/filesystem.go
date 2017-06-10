package entity_storage_filesystem

import (
	"path/filepath"

	"io/ioutil"

	"encoding/json"

	"encoding/base64"
	"os"

	"github.com/xiaonanln/goworld/common"
	"github.com/xiaonanln/goworld/gwlog"
	"github.com/xiaonanln/goworld/storage"
)

type FileSystemEntityStorage struct {
	directory string
}

func getFileName(name string, entityID common.EntityID) string {
	return name + "$" + base64.URLEncoding.EncodeToString([]byte(entityID))
}

func (ss *FileSystemEntityStorage) Write(name string, entityID common.EntityID, data interface{}) error {
	stringSaveFile := filepath.Join(ss.directory, getFileName(name, entityID))
	dataBytes, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	gwlog.Debug("Saving to file %s: %s", stringSaveFile, string(dataBytes))
	return ioutil.WriteFile(stringSaveFile, dataBytes, 0644)
}

func (ss *FileSystemEntityStorage) Read(name string, entityID common.EntityID) (interface{}, error) {
	stringSaveFile := filepath.Join(ss.directory, getFileName(name, entityID))
	dataBytes, err := ioutil.ReadFile(stringSaveFile)
	if err != nil {
		if os.IsNotExist(err) {
			// file not exist
			return nil, nil
		} else {
			return nil, err
		}
	}

	var data interface{}
	err = json.Unmarshal(dataBytes, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func newFileSystemEntityStorage(directory string) (*FileSystemEntityStorage, error) {
	if err := os.MkdirAll(directory, 0755); err != nil {
		return nil, err
	}

	return &FileSystemEntityStorage{
		directory: directory,
	}, nil
}

func OpenDirectory(directory string) (storage.EntityStorage, error) {
	return newFileSystemEntityStorage(directory)
}
