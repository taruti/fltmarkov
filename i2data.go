package main

import (
	"errors"
	"fmt"
	"os"

	mmap "github.com/edsrzf/mmap-go"
)

func RunI16(filename string, ncols int, nrows int, fun func([]int16, int, int) error) error {
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
	fs := unsafeBytesAsInt16([]byte(m))
	if len(fs) != ncols*nrows {
		return errors.New("Dimensions don't match to data size")
	}
	return fun(fs, ncols, nrows)
}

func DumpI16(fs []int16, ncols int, nrows int) error {
	rbase := 0
	var dmin, dmax, dnum, dsum, errD, min, max int32
	var dminC, dminR, dmaxC, dmaxR int
	var maxErr float64
	var arr [2000]int
	for row := 0; row < nrows; row++ {
		prev := int32(fs[rbase+0])
		for col := 0; col < ncols; col++ {
			d := int32(fs[rbase+col]) - prev
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

			/*
				ev := math.Abs(float64(half.From32(d).To32() - d))
				if ev > maxErr {
					maxErr = ev
					errD = d
				}
			*/
			aidx := int(d/10) + 500
			if aidx >= len(arr) {
				aidx = len(arr) - 1
			}
			if aidx < 0 {
				aidx = 0
			}
			arr[aidx]++
			prev = int32(fs[rbase+col])
			if prev < min {
				min = prev
			}
			if prev > max {
				max = prev
			}
		}
		rbase += ncols
	}

	fmt.Println("min", min)
	fmt.Println("max", max)
	fmt.Println("dmin", dmin, "at", dminC, dminR)
	fmt.Println("dmax", dmax, "at", dmaxC, dmaxR)
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
	/*
		fmt.Println("\n\narr")
		for i, v := range arr {
			if v > 1 {
				fmt.Printf("%5d %d\n", (i-500)*10, v)
			}
		}
	*/
	return nil
}
