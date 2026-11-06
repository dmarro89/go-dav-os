//go:build gccgo

package mem

import "unsafe"

func bitmapBytePtr(off uint64) *byte {
	// assumes identity mapping / paging off: physical == directly addressable pointer
	return (*byte)(unsafe.Pointer(uintptr(bitmapPhys) + uintptr(off)))
}
