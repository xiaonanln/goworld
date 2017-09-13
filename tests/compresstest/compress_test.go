package compresstest

import (
	"io"
	"math/rand"
	"os"
	"testing"

	"time"

	"compress/flate"

	"github.com/golang/snappy"
)

type randomFailWriter struct {
	io.Writer
}

func (rfw randomFailWriter) Write(b []byte) (int, error) {
	fv := rand.Float32()
	if fv < 0.1 {
		println("ErrShortWrite", fv)
		return 0, io.ErrShortWrite
	}
	return rfw.Writer.Write(b)
}

type Flusher interface {
	Flush() error
}

func TestFlateWriterSize(t *testing.T) {
	var err error
	f, err := os.Open("input.txt")
	if err != nil {
		panic(err)
	}

	all := []interface{}{}
	for i := 0; i < 5000; i++ {
		all = append(all, snappy.NewReader(f))
		all = append(all, snappy.NewBufferedWriter(f))
		all = append(all, flate.NewReader(f))
		w, _ := flate.NewWriter(f, flate.BestSpeed)
		all = append(all, w)
	}
	println("created")
	time.Sleep(time.Second * 10)
}

//func TestFlateCompressor(t *testing.T) {
//	var err error
//	f, err := os.Open("input.txt")
//	if err != nil {
//		panic(err)
//	}
//	of, err := os.OpenFile("output-flate.txt", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
//	if err != nil {
//		panic(err)
//	}
//
//	cw, err := flate.NewWriter(of, 1)
//	if err != nil {
//		panic(err)
//	}
//	testCompressor(t, f, cw)
//}
//
//func TestSnappyCompressor(t *testing.T) {
//	var err error
//	f, err := os.Open("input.txt")
//	if err != nil {
//		panic(err)
//	}
//	of, err := os.OpenFile("output-snappy.txt", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
//	if err != nil {
//		panic(err)
//	}
//
//	testCompressor(t, f, snappy.NewWriter(randomFailWriter{of}))
//}

//func TestLz4Compressor(t *testing.T) {
//	var err error
//	f, err := os.Open("input.txt")
//	if err != nil {
//		panic(err)
//	}
//	of, err := os.OpenFile("output-lz4.txt", os.O_CREATE|os.O_WRONLY, 0644)
//	if err != nil {
//		panic(err)
//	}
//
//	testCompressor(t, f, lz4.NewWriter(randomFailWriter{of}))
//}

func testCompressor(t *testing.T, r io.Reader, w io.Writer) {
	buf := make([]byte, 1024)
	for {
		n, err := r.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}

		if n > 0 {
			wbuf := buf[:n]
			for len(wbuf) > 0 {
				n, err := w.Write(wbuf)
				wbuf = wbuf[n:]
				if err != nil {
					println(n, err)
					w.(Flusher).Flush()
				}
			}
		}

	}
	w.(Flusher).Flush()
}
