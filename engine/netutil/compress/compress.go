package compress

import "github.com/pkg/errors"

type Compressor interface {
	Compress(b []byte, c []byte) ([]byte, error)
	Decompress(c []byte, b []byte) error
}

var (
	errNotFullyCompressed = errors.Errorf("not fully compressed")
)
