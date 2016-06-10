// +build ignore

package main

import (
	"math"
	"sort"
)

type NAME [SIZE]float64

func (r *NAME) Generate(data []float64) {
	const size = SIZE
	min := math.Inf(1)
	max := math.Inf(-1)
	for _, v := range data {
		if v < min {
			min = v
		}
		if v < max {
			max = v
		}
	}
	scale := max - min
	scalef := scale / float64(size)
	for i := range r {
		r[i] = float64(i) * scalef
	}
outer:
	for i := 0; i < 10*size; i++ {
		var counts [size]int
		var necounts [size]int
		for _, d := range data {
			bin, exact := r.linearBin(d)
			counts[bin]++
			if !exact {
				necounts[bin]++
			}
		}

		var maxv, maxi, nzero int
		for i, v := range necounts {
			if v > maxv {
				maxv = v
				maxi = i
			}
			if v == 0 {
				nzero++
			}
		}
		if nzero < 1 {
			break
		}
		for i, v := range counts {
			if v == 0 {
				dmax := 0.0
				var pfm float64
				for _, d := range data {
					bin, _ := r.linearBin(d)
					if bin == maxi {
						dist := math.Abs(r[maxi] - d)
						if dist > dmax {
							dmax = dist
							pfm = d
						}
					}
				}
				r[i] = pfm
				continue outer
			}
		}
	}
	sort.Float64s(r[:])
}

func (r *NAME) linearBin(d float64) (int, bool) {
	min := math.Inf(1)
	bin := 0
	for i, x := range r {
		v := math.Abs(d - x)
		if v < min {
			min = v
			bin = i
		}
		if v < 0.001 {
			return bin, true
		}
	}
	return bin, false
}
