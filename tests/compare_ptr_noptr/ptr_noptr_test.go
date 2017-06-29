package main

import "testing"

const (
	N = 100000
)

var (
	a = make(chan interface{}, 100)
)

func BenchmarkPtr(b *testing.B) {
	for i := 0; i < b.N; i++ {
		a <- &[N]byte{}
		_ = <-a
	}
}

func BenchmarkNoPtr(b *testing.B) {
	for i := 0; i < b.N; i++ {
		a <- [N]byte{}
		_ = <-a
	}
}
