package compress

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/xiaonanln/goworld/engine/gwlog"
)

type Compressor interface {
	Compress(b []byte, c []byte) ([]byte, error)
	Decompress(c []byte, b []byte) error
}

var (
	errNotFullyCompressed = errors.Errorf("not fully compressed")
)

func NewCompressor(compressFormat string) Compressor {
	compressFormat = strings.ToLower(compressFormat)
	if compressFormat == "snappy" {
		return NewSnappyCompressor()
	} else if compressFormat == "gwsnappy" {
		return NewGWSnappyCompressor()
	} else if compressFormat == "lz4" {
		return NewLz4Compressor()
	} else if compressFormat == "lzw" {
		return NewLzwCompressor()
	} else if compressFormat == "flate" {
		return NewFlateCompressor()
	} else {
		gwlog.Panicf("unknown compress format: %s", compressFormat)
		return nil
	}
}
