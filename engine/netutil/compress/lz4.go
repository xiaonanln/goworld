package compress

import (
	"os"

	"bytes"

	"github.com/pierrec/lz4"
	"github.com/xiaonanln/goworld/engine/netutil"
)

func NewLz4Compressor() Compressor {
	return &lz4Compressor{
		writer: lz4.NewWriter(os.Stdout),
		reader: lz4.NewReader(os.Stdin),
	}
}

type lz4Compressor struct {
	writer *lz4.Writer
	reader *lz4.Reader
}

func (sc *lz4Compressor) Compress(b []byte, c []byte) ([]byte, error) {
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

func (sc *lz4Compressor) Decompress(c []byte, b []byte) error {
	sc.reader.Reset(bytes.NewReader(c))
	return netutil.ReadAll(sc.reader, b)
}
