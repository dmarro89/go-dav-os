//go:build gccgo

package fs

import "github.com/dmarro89/go-dav-os/mem"

func pfaReady() bool {
	return mem.PFAReady()
}

func allocPage() uint64 {
	return mem.AllocPage()
}

func freePage(page uint64) bool {
	return mem.FreePage(page)
}
