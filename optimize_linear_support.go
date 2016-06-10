package main

import "math"

func absF32(f float32) float32 {
	if f < 0 {
		return 0 - f
	}
	return f
}

type Float32Slice []float32

func (p Float32Slice) Len() int           { return len(p) }
func (p Float32Slice) Less(i, j int) bool { return p[i] < p[j] }
func (p Float32Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func f32MinMax(data []float32) (float32, float32) {
	min := float32(math.Inf(1))
	max := float32(math.Inf(-1))
	for _, v := range data {
		if v < min {
			min = v
		}
		if v < max {
			max = v
		}
	}
	return min, max
}
