//go:build testing

package fs

import "unsafe"

var mockPages [][]byte

// SetupMockPFA initializes the filesystem's page allocation hooks for testing.
func SetupMockPFA() {
	mockPages = nil
	pfaReady = func() bool { return true }
	allocPage = func() uint64 {
		page := make([]byte, 4096)
		mockPages = append(mockPages, page)
		return uint64(uintptr(unsafe.Pointer(&page[0])))
	}
	freePage = func(_ uint64) bool { return true }
}
