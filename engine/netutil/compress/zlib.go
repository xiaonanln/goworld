package compress

import (
	"bytes"

	"compress/zlib"

	"io"

	"io/ioutil"

	"github.com/xiaonanln/goworld/engine/netutil"
)

func NewZlibCompressor() Compressor {
	fc := &zlibCompressor{}
	var err error
	fc.reader = newZlibReader()
	if err != nil {
		panic(err)
	}
	fc.writer, err = zlib.NewWriterLevel(ioutil.Discard, zlib.BestSpeed)
	if err != nil {
		panic(err)
	}
	return fc
}

func newZlibReader() (reader io.ReadCloser) {
	w := bytes.NewBuffer(nil)
	cw := zlib.NewWriter(w)
	cw.Write(nil)
	cw.Flush()

	reader, _ = zlib.NewReader(bytes.NewReader(w.Bytes()))
	return
}

type zlibCompressor struct {
	writer *zlib.Writer
	reader io.ReadCloser
}

func (fc *zlibCompressor) Compress(b []byte, c []byte) ([]byte, error) {
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

func (fc *zlibCompressor) Decompress(c []byte, b []byte) error {
	fc.reader.(zlib.Resetter).Reset(bytes.NewReader(c), nil)
	return netutil.ReadAll(fc.reader, b)
}
