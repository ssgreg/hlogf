package main

import (
	"unsafe"
)

// bytesToString is the simplest way to convert byte slice to a string without
// memory allocation.
func bytesToString(bs []byte) string {
	return *(*string)(unsafe.Pointer(&bs))
}

// atoi is the same as strconv.Atoi with two major differences.
// 	- atoi returns bool instead of an error that saves us from unnecessary
// 	allocation in case of error.
//	- atoi does not fail back to strconv.ParseInt.
func atoi(s string) (int, bool) {
	sLen := len(s)
	if intSize == 32 && (0 < sLen && sLen < 10) || intSize == 64 && (0 < sLen && sLen < 19) {
		n := 0
		for _, ch := range []byte(s) {
			ch -= '0'
			if ch > 9 {
				return 0, false
			}
			n = n*10 + int(ch)
		}

		return n, true
	}

	return 0, false
}

const intSize = 32 << (^uint(0) >> 63)
