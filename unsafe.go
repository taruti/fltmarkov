package main

import (
	"reflect"
	"unsafe"
)

func unsafeFloatsAsBytes(buf []float32) []byte {
	var bs []byte
	orig := (*reflect.SliceHeader)(unsafe.Pointer(&buf))
	dest := (*reflect.SliceHeader)(unsafe.Pointer(&bs))
	dest.Data = orig.Data
	dest.Cap = orig.Cap * 4
	dest.Len = orig.Len * 4
	return bs
}

func unsafeBytesAsFloats(buf []byte) []float32 {
	var bs []float32
	orig := (*reflect.SliceHeader)(unsafe.Pointer(&buf))
	dest := (*reflect.SliceHeader)(unsafe.Pointer(&bs))
	dest.Data = orig.Data
	dest.Cap = orig.Cap / 4
	dest.Len = orig.Len / 4
	return bs
}

func clearFloatSlice(fs []float32) {
	for i := range fs {
		fs[i] = 0
	}
}
