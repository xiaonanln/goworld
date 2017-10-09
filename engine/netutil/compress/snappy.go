package compress

import (
	"os"

	"bytes"

	"github.com/golang/snappy"
	"github.com/xiaonanln/goworld/engine/netutil"
)

func NewSnappyCompressor() Compressor {
	return &snappyCompressor{
		writer: snappy.NewWriter(os.Stdout),
		reader: snappy.NewReader(os.Stdin),
	}
}

type snappyCompressor struct {
	writer *snappy.Writer
	reader *snappy.Reader
}

func (sc *snappyCompressor) Compress(b []byte, c []byte) ([]byte, error) {
	wb := bytes.NewBuffer(c)
	sc.writer.Reset(wb)
	n, err := sc.writer.Write(b)
	if err != nil {
		return nil, err
	}
	if n != len(b) {
		return nil, errNotFullyCompressed
	}

	sc.writer.Flush()
	return wb.Bytes(), nil
}

func (sc *snappyCompressor) Decompress(c []byte, b []byte) error {
	sc.reader.Reset(bytes.NewReader(c))
	return netutil.ReadAll(sc.reader, b)
}
