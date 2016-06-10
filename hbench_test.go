package main

import (
	"math"
	"testing"

	h2 "github.com/foobaz/half"
	h1 "github.com/h2so5/half"
)

func BenchmarkH2So5Half(b *testing.B) {
	f := h1.NewFloat16(math.Pi)
	for i := 0; i < b.N; i++ {
		f.Float32()
	}
}

func BenchmarkFoobazHalf(b *testing.B) {
	f := h2.From64(math.Pi)
	for i := 0; i < b.N; i++ {
		f.To32()
	}
}
func BenchmarkFoobazHalf64(b *testing.B) {
	f := h2.From64(math.Pi)
	for i := 0; i < b.N; i++ {
		f.To64()
	}
}
