//go:build !gccgo

package fs

import "github.com/dmarro89/go-dav-os/mem"

var (
	pfaReady  = mem.PFAReady
	allocPage = mem.AllocPage
	freePage  = mem.FreePage
)
