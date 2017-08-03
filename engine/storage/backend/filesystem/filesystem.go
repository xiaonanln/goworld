package entity_storage_filesystem

import (
	"path/filepath"

	"io/ioutil"

	"encoding/json"

	"encoding/base64"
	"os"

	"strings"

	"github.com/xiaonanln/goworld/engine/common"
	"github.com/xiaonanln/goworld/engine/consts"
	"github.com/xiaonanln/goworld/engine/gwlog"
	. "github.com/xiaonanln/goworld/engine/storage/storage_common"
)

type FileSystemEntityStorage struct {
	directory string
}

func getFileName(name string, entityID common.EntityID) string {
	return name + "$" + base64.URLEncoding.EncodeToString([]byte(entityID))
}

func (es *FileSystemEntityStorage) getFilePath(typeName string, entityID common.EntityID) string {
	return filepath.Join(es.directory, getFileName(typeName, entityID))
}

func (es *FileSystemEntityStorage) Write(typeName string, entityID common.EntityID, data interface{}) error {
	stringSaveFile := es.getFilePath(typeName, entityID)
	dataBytes, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	if consts.DEBUG_SAVE_LOAD {
		gwlog.Debug("Saving to file %s: %s", stringSaveFile, string(dataBytes))
	}
	return ioutil.WriteFile(stringSaveFile, dataBytes, 0644)
}

func (es *FileSystemEntityStorage) Read(typeName string, entityID common.EntityID) (interface{}, error) {
	stringSaveFile := es.getFilePath(typeName, entityID)
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

func (es *FileSystemEntityStorage) Exists(typeName string, entityID common.EntityID) (exists bool, err error) {
	stringSaveFile := es.getFilePath(typeName, entityID)
	_, err = os.Stat(stringSaveFile)
	exists = err == nil || os.IsExist(err)
	return
}

func (es *FileSystemEntityStorage) List(typeName string) ([]common.EntityID, error) {
	prefix := typeName + "$"
	pat := filepath.Join(es.directory, prefix+"*")
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

func (es *FileSystemEntityStorage) Close() {
	// need to do nothing
}

func (es *FileSystemEntityStorage) IsEOF(err error) bool {
	return false
}

func OpenDirectory(directory string) (EntityStorage, error) {
	if err := os.MkdirAll(directory, 0755); err != nil {
		return nil, err
	}

	return &FileSystemEntityStorage{
		directory: directory,
	}, nil
}
