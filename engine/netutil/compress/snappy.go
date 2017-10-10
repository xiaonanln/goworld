package compress

import (
	"os"

	"bytes"

	"github.com/xiaonanln/goworld/engine/gwioutil"
	"github.com/xiaonanln/goworld/engine/lib/gwsnappy"
)

func NewSnappyCompressor() Compressor {
	return &snappyCompressor{
		writer: gwsnappy.NewWriter(os.Stdout),
		reader: gwsnappy.NewReader(os.Stdin),
	}
}

type snappyCompressor struct {
	writer *gwsnappy.Writer
	reader *gwsnappy.Reader
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
	return gwioutil.ReadAll(sc.reader, b)
}
