package entitystoragefilesystem

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
	"github.com/xiaonanln/goworld/engine/storage/storage_common"
)

// FileSystemEntityStorage is an implementation of Entity Storage using filesystem
type FileSystemEntityStorage struct {
	directory string
}

func getFileName(name string, entityID common.EntityID) string {
	return name + "$" + base64.URLEncoding.EncodeToString([]byte(entityID))
}

func (es *FileSystemEntityStorage) getFilePath(typeName string, entityID common.EntityID) string {
	return filepath.Join(es.directory, getFileName(typeName, entityID))
}

// Write writes entity data to entity storage
func (es *FileSystemEntityStorage) Write(typeName string, entityID common.EntityID, data interface{}) error {
	stringSaveFile := es.getFilePath(typeName, entityID)
	dataBytes, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	if consts.DEBUG_SAVE_LOAD {
		gwlog.Debugf("Saving to file %s: %s", stringSaveFile, string(dataBytes))
	}
	return ioutil.WriteFile(stringSaveFile, dataBytes, 0644)
}

// Read reads entity data from entity storage
func (es *FileSystemEntityStorage) Read(typeName string, entityID common.EntityID) (interface{}, error) {
	stringSaveFile := es.getFilePath(typeName, entityID)
	dataBytes, err := ioutil.ReadFile(stringSaveFile)
	if err != nil {
		if os.IsNotExist(err) {
			// file not exist
			return nil, nil
		}
		return nil, err
	}

	var data interface{}
	err = json.Unmarshal(dataBytes, &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// Exists checks if entity is in entity storage
func (es *FileSystemEntityStorage) Exists(typeName string, entityID common.EntityID) (exists bool, err error) {
	stringSaveFile := es.getFilePath(typeName, entityID)
	_, err = os.Stat(stringSaveFile)
	exists = err == nil || os.IsExist(err)
	return
}

// List retrives all entity IDs in entity storage of specified type
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
			gwlog.Errorf("invalid file: %s", fpath)
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

// Close the entity storage
func (es *FileSystemEntityStorage) Close() {
	// need to do nothing
}

// IsEOF check if the error is an EOF error
func (es *FileSystemEntityStorage) IsEOF(err error) bool {
	return false
}

// OpenDirectory opens the directory as filesystem entity storage
func OpenDirectory(directory string) (storagecommon.EntityStorage, error) {
	if err := os.MkdirAll(directory, 0755); err != nil {
		return nil, err
	}

	return &FileSystemEntityStorage{
		directory: directory,
	}, nil
}
