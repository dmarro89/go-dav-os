//go:build gccgo

package fs

import "github.com/dmarro89/go-dav-os/mem"

func pfaReady() bool {
	if mockPFAActive {
		return true
	}
	return mem.PFAReady()
}

func allocPage() uint64 {
	if mockPFAActive && mockAllocPageFn != nil {
		return mockAllocPageFn()
	}
	return mem.AllocPage()
}

func freePage(page uint64) bool {
	if mockPFAActive {
		return true
	}
	return mem.FreePage(page)
}
