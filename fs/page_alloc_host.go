//go:build !gccgo

package fs

import "github.com/dmarro89/go-dav-os/mem"

var (
	pfaReady = func() bool {
		if mockPFAActive {
			return true
		}
		return mem.PFAReady()
	}
	allocPage = func() uint64 {
		if mockPFAActive && mockAllocPageFn != nil {
			return mockAllocPageFn()
		}
		return mem.AllocPage()
	}
	freePage = func(page uint64) bool {
		if mockPFAActive {
			return true
		}
		return mem.FreePage(page)
	}
)
