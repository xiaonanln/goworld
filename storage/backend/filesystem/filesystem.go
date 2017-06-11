package entity_storage_filesystem

import (
	"path/filepath"

	"io/ioutil"

	"encoding/json"

	"encoding/base64"
	"os"

	"strings"

	"github.com/xiaonanln/goworld/common"
	"github.com/xiaonanln/goworld/gwlog"
	"github.com/xiaonanln/goworld/storage/common"
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

func (ss *FileSystemEntityStorage) Read(typeName string, entityID common.EntityID) (interface{}, error) {
	stringSaveFile := filepath.Join(ss.directory, getFileName(typeName, entityID))
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

func (ss *FileSystemEntityStorage) List(typeName string) ([]common.EntityID, error) {
	prefix := typeName + "$"
	pat := filepath.Join(ss.directory, prefix+"*")
	files, err := filepath.Glob(pat)
	if err != nil {
		return nil, err
	}
	res := make([]common.EntityID, 0, len(files))
	prefixLen := len(prefix)
	for _, fpath := range files {
		_, fn := filepath.Split(fpath)
		if !strings.HasPrefix(fn, prefix) {
			gwlog.Error("invalid file: %s", fpath)
		}
		idbytes, err := base64.URLEncoding.DecodeString(fn[prefixLen:])
		if err != nil {
			gwlog.TraceError("fail to parse file %s", fpath)
			continue
		}

		res = append(res, common.MustEntityID(string(idbytes)))
	}
	return res, nil
}

func newFileSystemEntityStorage(directory string) (*FileSystemEntityStorage, error) {
	if err := os.MkdirAll(directory, 0755); err != nil {
		return nil, err
	}

	return &FileSystemEntityStorage{
		directory: directory,
	}, nil
}

func OpenDirectory(directory string) (storage_common.EntityStorage, error) {
	return newFileSystemEntityStorage(directory)
}
