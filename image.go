package main

import (
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"log"
	"math"
	"os"
)

type flc struct {
	fs           []float32
	ncols, nrows int
	scale        float64
}

func (f *flc) ColorModel() color.Model { return color.RGBAModel }
func (f *flc) Bounds() image.Rectangle { return image.Rect(0, 0, f.ncols, f.nrows) }
func (i *flc) At(x, y int) color.Color {
	f := i.fs[y*i.ncols+x]
	if f < 0 {
		return color.RGBA{0, 0, 0xFF, 00}
	}
	t := int(math.Log(float64(f)) * i.scale)
	if t > 0xFF {
		t = 0xFF
	}
	b := byte(t)
	return color.RGBA{b, b, b, 0}
}

func AsImage(fs []float32, ncols int, nrows int) image.Image {
	max := float32(0.0)
	for _, v := range fs {
		if v > max {
			max = v
		}
	}
	s := 0xff / math.Log(float64(max))
	return &flc{fs, ncols, nrows, s}
}

func WriteJPEG(fn string, img image.Image) error {
	f, e := os.Create(fn)
	if e != nil {
		return e
	}
	defer f.Close()
	return jpeg.Encode(f, img, nil)
}

type AsJPEG struct {
	OutFile string
	Scale   bool
}

func (a AsJPEG) Work(fs []float32, ncols int, nrows int) error {
	if a.Scale {
		oc, or := ncols, nrows
		fs, ncols, nrows = DownScaleN(fs, ncols, nrows, 8)
		log.Println("Scaled", oc, "x", or, "=>", ncols, "x", nrows)
	}
	return WriteJPEG(a.OutFile, AsImage(fs, ncols, nrows))
}

func WritePNG(fn string, img image.Image) error {
	f, e := os.Create(fn)
	if e != nil {
		return e
	}
	defer f.Close()
	return png.Encode(f, img)
}
