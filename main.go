package main

import (
	"bufio"
	"errors"
	"fmt"
	"image"
	"io"
	"math"
	"math/rand"
	"os"
	"runtime"

	"github.com/foobaz/half"

	mmap "github.com/edsrzf/mmap-go"
)

func Run(filename string, ncols int, nrows int, fun func([]float32, int, int) error) error {
	f, e := os.Open(filename)
	if e != nil {
		return e
	}
	defer f.Close()
	m, e := mmap.Map(f, mmap.RDONLY, 0)
	if e != nil {
		return e
	}
	defer m.Unmap()
	fs := unsafeBytesAsFloats([]byte(m))
	if len(fs) != ncols*nrows {
		return errors.New("Dimensions don't match to data size")
	}
	return fun(fs, ncols, nrows)
}

func Dump(fs []float32, ncols int, nrows int) error {
	rbase := 0
	var dmin, dmax, dnum, dsum, errD float32
	var dminC, dminR, dmaxC, dmaxR int
	var maxErr float64
	var arr [2000]int
	for row := 0; row < nrows; row++ {
		prev := fs[rbase+0]
		for col := 0; col < ncols; col++ {
			d := fs[rbase+col] - prev
			if d < dmin {
				dmin = d
				dminC = col
				dminR = row
			}
			if d > dmax {
				dmax = d
				dmaxC = col
				dmaxR = row
			}
			dnum++
			dsum += d

			ev := math.Abs(float64(half.From32(d).To32() - d))
			if ev > maxErr {
				maxErr = ev
				errD = d
			}

			aidx := int(d/10) + 500
			if aidx >= len(arr) {
				aidx = len(arr) - 1
			}
			if aidx < 0 {
				aidx = 0
			}
			arr[aidx]++
			prev = fs[rbase+col]
		}
		rbase += ncols
	}

	fmt.Println("min", dmin, "at", dminC, dminR)
	fmt.Println("max", dmax, "at", dmaxC, dmaxR)
	fmt.Println("num", dnum)
	fmt.Println("sum", dsum)
	fmt.Println("avg", dsum/dnum)
	fmt.Println("err16", maxErr, "from", errD)

	fmt.Println("\n\nmin")
	for c := dminC - 4; c < dminC+3; c++ {
		fmt.Print(fs[(dminR*ncols)+c], " ")
	}

	fmt.Println("\n\nmax")
	for c := dmaxC - 4; c < dmaxC+3; c++ {
		fmt.Print(fs[(dmaxR*ncols)+c], " ")
	}
	fmt.Println()

	for i, v := range arr {
		if v > 1 {
			fmt.Printf("%5d %d\n", (i-500)*10, v)
		}
	}

	return nil
}

const (
	nslotBits   = 4
	nslots      = 1 << nslotBits
	numpixels   = 4
	npixelSlots = 1 << (numpixels * nslotBits)
)

type Slot [nslots]half.Float16

type BMarkov struct {
	arr [npixelSlots]Slot
}

func (bm *BMarkov) WriteTo(w io.Writer) (int, error) {
	mw := bufio.NewWriter(w)
	bs := make([]byte, nslots*2)
	var n, tot int
	var e error
	for ai := range bm.arr {
		for j, v := range bm.arr[ai] {
			v.PutLittleEndian(bs[j*2:])
		}
		n, e = mw.Write(bs)
		tot += n
		if e != nil {
			return tot, e
		}
	}
	e = mw.Flush()
	return tot, e
}

func (bm *BMarkov) Generate(ncols int, nrows int) []float32 {
	fs := make([]float32, ncols*nrows)
	rbase := 0
	prev := 0.0
	for row := 0; row < 2; row++ {
		for col := 0; col < ncols; col++ {
			prev = prev + ((rand.Float64() - 0.5) * 100)
			fs[rbase+col] = float32(prev)
		}
		rbase += ncols
	}
	for row := 2; row < nrows; row++ {
		for col := 0; col < 2; col++ {
			prev = prev + ((rand.Float64() - 0.5) * 100)
			if prev < 1 {
				prev = 1
			}
			fs[rbase+col] = float32(prev)
		}
		rbase += ncols
	}
	rbase = 2 * ncols
	for row := 2; row < nrows; row++ {
		for col := 2; col < ncols-1; col++ {
			p1 := fs[rbase+col-1]
			p2 := fs[(rbase-(ncols-ncols))+col-2]
			p3 := fs[(rbase-ncols)+col]
			p4 := fs[(rbase-ncols)+col+1]
			slotIdx := bmIndex(p1, p2, p3, p4)
			rv := rand.Float64()
			ci := 0
		selectLoop:
			for i, v := range bm.arr[slotIdx] {
				rv = rv - v.To64()
				if rv <= 0 {
					ci = i
					break selectLoop
				}
			}
			if ci > 8 {
				ci -= 8
			}
			v := math.Exp(float64(ci))
			if ci > 8 {
				v = 0 - v
			}
			fs[rbase+col] = float32(v)
		}
		rbase += ncols
	}
	return fs
}

type LearnMarkov struct {
	*BMarkov
	image.Rectangle
	OutFile string
}

func (lm *LearnMarkov) Work(fs []float32, ncols int, nrows int) error {
	fs, ncols, nrows = DownScaleN(fs, ncols, nrows, 8)
	rbase := 2 * ncols
	for row := 2; row < nrows; row++ {
		for col := 2; col < ncols; col++ {
			p1 := fs[rbase+col-1]
			p2 := fs[(rbase-(ncols-ncols))+col-2]
			p3 := fs[(rbase-ncols)+col]
			p4 := fs[(rbase-ncols)+col+1]
			cur := fs[rbase+col]
			if p1 < 0 || p2 < 0 || p3 < 0 || p4 < 0 || cur < 0 {
				continue
			}
			slotIdx := bmIndex(p1, p2, p3, p4)
			fiv := fi(cur)
			//			if slotIdx >= len(lm.BMarkov.arr) || fiv >= nslots {
			//			log.Println("slot", row, col, slotIdx, fi(cur))
			//			}
			lm.BMarkov.arr[slotIdx][fiv]++
		}
		rbase += ncols
	}
	for i := range lm.BMarkov.arr {
		sum := 0.0
		for _, v := range lm.BMarkov.arr[i] {
			sum += v.To64()
		}
		if sum > 0 {
			for j, v := range lm.BMarkov.arr[i] {
				lm.BMarkov.arr[i][j] = half.From64(v.To64() / sum)
			}
		} else {
			c := half.From64(1.0 / nslots)
			for j := range lm.BMarkov.arr[i] {
				lm.BMarkov.arr[i][j] = c
			}
		}
	}
	if lm.OutFile != "" {
		f, e := os.Create(lm.OutFile)
		if e != nil {
			return e
		}
		defer f.Close()
		lm.BMarkov.WriteTo(f)
	}
	if true {
		c, r := 1000, 1000
		fs := lm.BMarkov.Generate(c, r)
		AsJPEG{"/tmp/m1.jpg", false}.Work(fs, c, r)
	}
	return nil
}

func fi(f float32) int {
	if f < 0.1 {
		return 0
	}
	v := int(math.Log(math.Abs(float64(f))) + 0.5)
	if v <= 0 {
		return 0
	}
	if v > 7 {
		v = 7
	}
	if f < 0 {
		v |= 8
	}
	return v
}
func bmIndex(p1, p2, p3, p4 float32) int {
	return fi(p1) | (fi(p2) << 4) | (fi(p3) << 8) | (fi(p4) << 12)
}

type Worker interface {
	Work([]float32, int, int) error
}

func Ruw(fn string, ncols int, nrows int, w Worker) error {
	return Run(fn, ncols, nrows, w.Work)
}

func main() {
	//	Run("/tmp/data", 21601, 10801, Dump)
	//	Ruw("/tmp/data", 21601, 10801, AsJPEG{"/tmp/out.jpg", true})
	//Ruw("/tmp/data", 21601, 10801, &LearnMarkov{new(BMarkov), image.Rect(0, 0, 1000, 1000), "/tmp/out.mkov"})
	Ruw("/tmp/data", 21601, 10801, BuildClassifier{})
}
