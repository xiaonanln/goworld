package compress

import (
	"math/rand"
	"testing"
)

func BenchmarkSnappyCompressor(b *testing.B) {
	benchmarkCompressor(b, NewSnappyCompressor())
}

func BenchmarkLz4Compressor(b *testing.B) {
	benchmarkCompressor(b, NewLz4Compressor())
}

func BenchmarkLzwCompressor(b *testing.B) {
	benchmarkCompressor(b, NewLzwCompressor())
}

func BenchmarkFlateCompressor(b *testing.B) {
	benchmarkCompressor(b, NewFlateCompressor())
}

func BenchmarkZLibCompressor(b *testing.B) {
	benchmarkCompressor(b, NewZlibCompressor())
}

func benchmarkCompressor(b_ *testing.B, cr Compressor) {
	dataSize := 100 * 1024
	b := make([]byte, dataSize)
	for j := 0; j < dataSize; j++ {
		b[j] = byte(97 + rand.Intn(3))
	}

	for n := 0; n < b_.N; n++ {
		var c []byte
		var err error
		if c, err = cr.Compress(b, c); err != nil {
			panic(err)
		}

		rb := make([]byte, len(b))
		if err = cr.Decompress(c, rb); err != nil {
			panic(err)
		}

		dataSize = dataSize * 2
	}
}

func TestSnappyCompressor(t *testing.T) {
	testCompressor(t, NewSnappyCompressor())
}

func TestLz4Compressor(t *testing.T) {
	testCompressor(t, NewLz4Compressor())
}

func TestFlateCompressor(t *testing.T) {
	testCompressor(t, NewFlateCompressor())
}

func TestZlibCompressor(t *testing.T) {
	testCompressor(t, NewZlibCompressor())
}

func TestLzwCompressor(t *testing.T) {
	testCompressor(t, NewLzwCompressor())
}

func testCompressor(t *testing.T, cr Compressor) {
	dataSize := 10 * 1024
	for i := 0; i < 10; i++ {
		b := make([]byte, dataSize)
		for j := 0; j < dataSize; j++ {
			b[j] = byte(97 + rand.Intn(10))
		}
		//t.Logf("Compressing %s", string(b))

		var c []byte
		var err error
		if c, err = cr.Compress(b, c); err != nil {
			t.Fatal(err)
		}

		t.Logf("original size is %d, compressed size is %d (%d%%)", len(b), len(c), len(c)*100/len(b))

		rb := make([]byte, len(b))
		if err = cr.Decompress(c, rb); err != nil {
			t.Fatal(err)
		}

		if len(rb) != len(b) {
			t.Errorf("original data size is %d, but restore data size is %d", len(b), len(rb))
		}

		if string(rb) != string(b) {
			t.Errorf("original data and restored data mismatch", len(b), len(rb))
		}

		dataSize = dataSize * 2
	}
}
