package main

import (
	"sync/atomic"
	"testing"
)

func BenchmarkAtomicLoadInt32(b *testing.B) {
	var v int32
	for i := 0; i < b.N; i++ {
		atomic.LoadInt32(&v)
	}
}

func BenchmarkAtomicLoadUint32(b *testing.B) {
	var v uint32
	for i := 0; i < b.N; i++ {
		atomic.LoadUint32(&v)
	}
}

func BenchmarkAtomicLoadInt64(b *testing.B) {
	var v int64
	for i := 0; i < b.N; i++ {
		atomic.LoadInt64(&v)
	}
}

func BenchmarkAtomicLoadUint64(b *testing.B) {
	var v uint64
	for i := 0; i < b.N; i++ {
		atomic.LoadUint64(&v)
	}
}
