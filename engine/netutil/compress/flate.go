package compress

import (
	"os"

	"bytes"

	"compress/flate"

	"io"

	"io/ioutil"

	"github.com/xiaonanln/goworld/engine/netutil"
)

func NewFlateCompressor() Compressor {
	fc := &flateCompressor{
		reader: flate.NewReader(os.Stdin),
	}
	var err error
	fc.writer, err = flate.NewWriter(ioutil.Discard, flate.BestSpeed)
	if err != nil {
		panic(err)
	}
	return fc
}

type flateCompressor struct {
	writer *flate.Writer
	reader io.ReadCloser
}

func (fc *flateCompressor) Compress(b []byte, c []byte) ([]byte, error) {
	wb := bytes.NewBuffer(c)
	fc.writer.Reset(wb)
	n, err := fc.writer.Write(b)
	if err != nil {
		return nil, err
	}
	if n != len(b) {
		return nil, errNotFullyCompressed
	}

	fc.writer.Flush()
	return wb.Bytes(), nil
}

func (fc *flateCompressor) Decompress(c []byte, b []byte) error {
	fc.reader.(flate.Resetter).Reset(bytes.NewReader(c), nil)
	return netutil.ReadAll(fc.reader, b)
}
