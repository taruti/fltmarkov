package main

import (
	"fmt"
	"image"
	"image/color"
	"image/color/palette"
	"image/draw"
	"log"
	"math"
)

func intMax(a, b int) int {
	if a > b {
		return a
	}
	return b
}
func intMin(a, b int) int {
	if a < b {
		return a
	}
	return b
}

type FS struct {
	Data  []float32
	NCols int
	NRows int
}

type PartFeature struct{ Avg, StdDev float64 }

func partFeatures(f *FS, col int, row int, d int) PartFeature {
	fs, ncols, nrows := f.Data, f.NCols, f.NRows
	rowMax := intMin(nrows, row+d)
	colMax := intMin(ncols, col+d)
	var sum float64
	var n int
	for row := row; row < rowMax; row++ {
		for col := col; col < colMax; col++ {
			v := (float64(fs[row*ncols+col]))
			if v >= 0 {
				sum += v
				n++
			}
		}
	}
	if n == 0 {
		return PartFeature{}
	}
	avg := sum / float64(n)
	sum = 0
	for row := row; row < rowMax; row++ {
		for col := col; col < colMax; col++ {
			v := (float64(fs[row*ncols+col]))
			if v >= 0 {
				diff := math.Abs(v - avg)
				sum += diff * diff
			}
		}
	}

	return PartFeature{avg, math.Sqrt(sum / float64(n))}
}

func buildFeatureClassifier(f *FS) {
	d := 100
	var fs []PartFeature

	log.Println("Fetching features d", d)
	for row := 0; row < f.NRows; row += d {
		for col := 0; col < f.NCols; col += d {
			fs = append(fs, partFeatures(f, col, row, d))
		}
	}

	log.Println("Got", len(fs), "features")

	var asum, amin, amax, ssum, smin, smax float64
	for _, v := range fs {
		asum += v.Avg
		if v.Avg < amin {
			amin = v.Avg
		}
		if v.Avg > amax {
			amax = v.Avg
		}
		ssum += v.StdDev
		if v.StdDev < smin {
			smin = v.StdDev
		}
		if v.StdDev > smax {
			smax = v.StdDev
		}
	}
	log.Println("Average min/avg/max", amin, asum/float64(len(fs)), amax)
	log.Println("StdDev  min/avg/max", smin, ssum/float64(len(fs)), smax)
	ascale := amax - amin
	sscale := smax - smin
	dscale := ascale / sscale

	/// LINEAR
	as := make([]float32, len(fs))
	ss := make([]float32, len(fs))
	for i, f := range fs {
		as[i] = float32(f.Avg)
		ss[i] = float32(f.StdDev)
	}
	log.Println("Starting linear optimization")
	var aL, sL OptimizeLinear256
	ch := make(chan struct{})
	go func() {
		sL.Generate(ss)
		ch <- struct{}{}
	}()
	aL.Generate(as)
	<-ch
	log.Println("Done linear optimization")
	clim := (f.NCols + d - 1) / d
	rlim := (f.NRows + d - 1) / d
	dst := image.NewRGBA(image.Rect(0, 0, clim, rlim))
	i := 0
	log.Println("Starting to write image")
	for row := 0; row < rlim; row++ {
		for col := 0; col < clim; col++ {
			f := fs[i]
			c1 := fidX(f.Avg, &aL)
			c2 := fidX(f.StdDev, &sL)
			dst.SetRGBA(col, row, color.RGBA{c1, c1, c2, 0x0})
			i++
		}
	}
	log.Println("Compressing")
	WriteJPEG("/tmp/l.jpg", dst)
	log.Println("Done")

	gdst := image.NewGray(image.Rect(0, 0, clim, rlim))
	i = 0
	for row := 0; row < rlim; row++ {
		for col := 0; col < clim; col++ {
			f := fs[i]
			c1 := fidX(f.Avg, &aL)
			//			dst.SetRGBA(col, row, color.RGBA{c1, c1, c1, 0x0})
			gdst.SetGray(col, row, color.Gray{c1})
			i++
		}
	}
	//	WriteJPEG("/tmp/lh.jpg", dst)
	//WriteJPEG("/tmp/lhg.jpg", gdst)
	WritePNG("/tmp/lhg.png", gdst)
	//	dst = image.NewRGBA(image.Rect(0, 0, clim, rlim))
	i = 0
	for row := 0; row < rlim; row++ {
		for col := 0; col < clim; col++ {
			f := fs[i]
			c2 := fidX(f.StdDev, &sL)
			//			dst.SetRGBA(col, row, color.RGBA{c2, c2, c2, 0x0})
			gdst.SetGray(col, row, color.Gray{c2})
			i++
		}
	}
	WritePNG("/tmp/lsg.png", gdst)

	// Comb Lin
	log.Println("Starting linear optimization 64")
	var aLL, sLL OptimizeLinear64
	go func() {
		sLL.Generate(ss)
		ch <- struct{}{}
	}()
	aLL.Generate(as)
	<-ch
	log.Println("Done linear optimization 64")

	i = 0
	for row := 0; row < rlim; row++ {
		for col := 0; col < clim; col++ {
			f := fs[i]
			c1 := byte(aLL.Find64(f.Avg) * 4)
			c2 := byte(sLL.Find64(f.StdDev) * 4)
			dst.SetRGBA(col, row, color.RGBA{c1, c1, c2, 0xFF})
			i++
		}
	}
	WriteJPEG("/tmp/lc64.jpg", dst)
	WritePNG("/tmp/lc64.png", dst)

	return

	/// AUTOBIN

	c := &Cand{}
	for i := 0; i < 16; i++ {
		for j := 0; j < 16; j++ {
			c[i*16+j] = PartFeature{float64(i) * ascale, float64(j) * sscale}
		}
	}

	c.Optimize(fs, dscale)

	i = 0
	dst = image.NewRGBA(image.Rect(0, 0, f.NCols, f.NRows))
	for row := 0; row < f.NRows; row += d {
		for col := 0; col < f.NCols; col += d {
			f := fs[i]
			bin := c.Fit(&f, dscale)
			r, g, b, _ := palette.Plan9[bin].RGBA()
			simg := &SImage{d, color.RGBA{byte(r), byte(g), byte(b), 0x80}}
			draw.Draw(dst, image.Rect(col, row, col+d, row+d), simg, image.Point{0, 0}, draw.Over)
			i++
		}
	}
	WriteJPEG("/tmp/c.jpg", dst)
}

func fidX16(f float64, farr *OptimizeLinear16) byte {
	// Linear search is faster than binary search here...
	x := float32(f)
	for i, v := range *farr {
		if v >= x {
			return byte(i)
		}
	}
	return 15
}
func fidX(f float64, farr *OptimizeLinear256) byte {
	return byte(farr.Find64(f))
}

type SImage struct {
	D int
	C color.RGBA
}

func (f *SImage) ColorModel() color.Model { return color.RGBAModel }
func (f *SImage) Bounds() image.Rectangle { return image.Rect(0, 0, f.D, f.D) }
func (f *SImage) At(x, y int) color.Color { return f.C }

type Cand [0x100]PartFeature

func (c *Cand) Optimize(fs []PartFeature, ds float64) {
	log.Println("Starting to optimize candidates")
outer:
	for x := 0; x < 1000; x++ {
		var counts [0x100]int
		for _, f := range fs {
			counts[c.Fit(&f, ds)]++
		}
		var maxv, maxi, nzero int
		for i, v := range counts {
			if v > maxv && i != 0 {
				maxv = v
				maxi = i
			}
			if v == 0 {
				nzero++
			}
		}
		if nzero < 2 {
			break
		}
		fmt.Println("Optimizing nzero", nzero, maxv, "@", maxi, counts)
		//		for i, v := range counts {
		//			if v == 0 {
		//				c[i] = PartFeature{c[maxi].Avg + (rand.Float64()-0.5)*20, c[maxi].StdDev + (rand.Float64() - 0.5)}
		//			}
		for i, v := range counts {
			if v == 0 {
				fmt.Println("v is zero at", i)
				k := 0
				dmax := 0.0
				var pfm PartFeature
				for _, f := range fs {
					if c.Fit(&f, ds) == maxi {
						d := c[maxi].Distance(&f, ds)
						if d > dmax {
							dmax = d
							pfm = f
						}
						k++
					}
				}
				log.Println("OptMove", c[i], "=>", pfm)
				c[i] = pfm
				continue outer
			}
		}
	}
	log.Println("Done optimizing candidates")
}

func (c *Cand) Fit(o *PartFeature, dscale float64) int {
	var bestFit int
	bestD := math.Inf(1)
	for i, pf := range c {
		d := pf.Distance(o, dscale)
		if d < bestD {
			bestD = d
			bestFit = i
		}
	}
	return int(byte(bestFit))
}

func (p *PartFeature) Distance(o *PartFeature, devScale float64) float64 {
	//	return math.Abs(p.StdDev - o.StdDev)
	avgSum := (p.Avg - o.Avg)
	sdvSum := (p.StdDev - o.StdDev) * devScale
	return math.Sqrt(avgSum*avgSum + sdvSum*sdvSum)
}

type BuildClassifier struct{}

func (BuildClassifier) Work(fs []float32, ncols int, nrows int) error {
	buildFeatureClassifier(&FS{fs, ncols, nrows})
	return nil
}
