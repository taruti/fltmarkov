package main

import (
	"math/rand"
	"testing"
)

var rfs []float32

func init() {
	rfs = make([]float32, 20000)
	for i := range rfs {
		rfs[i] = float32(rand.Float64()*100 + rand.Float64()*10)
	}
}

func BenchmarkOptimizeLinear16(b *testing.B) {
	var lo OptimizeLinear16
	for i := 0; i < b.N; i++ {
		lo.Generate(rfs)
	}
}

func BenchmarkOptimizeLinear256(b *testing.B) {
	var lo OptimizeLinear256
	for i := 0; i < b.N; i++ {
		lo.Generate(rfs)
	}
}
