package stringx

import (
	"reflect"
	"unsafe"
)

// S2b converts a string to a byte slice.
//
// This is a shallow copy, means that the returned byte slice reuse
// the underlying array in string, so you can't change the returned
// byte slice in any situations.
func S2b(s string) (b []byte) {
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh.Data = sh.Data
	bh.Len = sh.Len
	bh.Cap = sh.Len
	return b
}

// B2s converts a byte slice to a string.
//
// This is a shallow copy, means that the returned string reuse the
// underlying array in byte slice, it's your responsibility to keep
// the input byte slice survive until you don't access the string anymore.
func B2s(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
