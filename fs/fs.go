package fs

import (
	"unsafe"

	"github.com/dmarro89/go-dav-os/mem"
)

const (
	maxFiles = 32
	maxName  = 16
	pageSize = 4096
)

// Memory operations vars to allow mocking in tests
var (
	memAlloc    = mem.AllocPage
	memFree     = mem.FreePage
	memPFAReady = mem.PFAReady
	// memWrite writes data to a physical page.
	// We abstract this to allow tests on 64-bit hosts where casting physical
	// addresses (uint32) to pointers (uintptr) would otherwise crash or be invalid.
	memWrite = func(pagePhys uint32, data *byte, len uint32) {
		dstBase := uintptr(pagePhys)
		srcBase := uintptr(unsafe.Pointer(data))
		for i := uint32(0); i < len; i++ {
			*(*byte)(unsafe.Pointer(dstBase + uintptr(i))) =
				*(*byte)(unsafe.Pointer(srcBase + uintptr(i)))
		}
	}
)

type fileEntry struct {
	used    bool
	nameLen uint8
	name    [maxName]byte
	size    uint32
	page    uint32 // physical address of the page
}

var files [maxFiles]fileEntry

// Init resets the in-memory filesystem table
func Init() {
	for i := 0; i < maxFiles; i++ {
		files[i].used = false
		files[i].nameLen = 0
		files[i].size = 0
		files[i].page = 0
	}
}

func MaxFiles() int { return maxFiles }

// Entry returns metadata for the i-th slot
// iterate i=0..MaxFiles()-1 and check if used
func Entry(i int) (used bool, name *[maxName]byte, nameLen int, size uint32, page uint32) {
	if i < 0 || i >= maxFiles {
		return false, nil, 0, 0, 0
	}
	e := &files[i]
	return e.used, &e.name, int(e.nameLen), e.size, e.page
}

// Lookup finds a file by name and returns its backing page + size.
func Lookup(name *[maxName]byte, nameLen int) (page uint32, size uint32, ok bool) {
	idx := findByName(name, nameLen)
	if idx < 0 {
		return 0, 0, false
	}
	e := &files[idx]
	return e.page, e.size, true
}

// Write creates or overwrites a file
// data is copied into the file backing page
func Write(name *[maxName]byte, nameLen int, data *byte, dataLen uint32) bool {
	if nameLen <= 0 || nameLen > maxName {
		return false
	}
	if dataLen > pageSize {
		dataLen = pageSize
	}

	idx := findByName(name, nameLen)
	if idx < 0 {
		idx = findFreeSlot()
		if idx < 0 {
			return false
		}
	}

	e := &files[idx]

	// allocate a page if this is a new file
	if !e.used {
		if !memPFAReady() {
			return false
		}
		p := memAlloc()
		if p == 0 {
			return false
		}
		e.used = true
		e.page = p
		copyName(e, name, nameLen)
	}

	// copy data into the backing page (physical memory)
	memWrite(e.page, data, dataLen)
	e.size = dataLen
	return true
}

// Remove deletes a file and frees its backing page
func Remove(name *[maxName]byte, nameLen int) bool {
	idx := findByName(name, nameLen)
	if idx < 0 {
		return false
	}

	e := &files[idx]
	if e.used && e.page != 0 {
		memFree(e.page)
	}

	e.used = false
	e.nameLen = 0
	e.size = 0
	e.page = 0
	return true
}

func findFreeSlot() int {
	for i := 0; i < maxFiles; i++ {
		if !files[i].used {
			return i
		}
	}
	return -1
}

func findByName(name *[maxName]byte, nameLen int) int {
	if nameLen <= 0 || nameLen > maxName {
		return -1
	}
	for i := 0; i < maxFiles; i++ {
		e := &files[i]
		if !e.used || int(e.nameLen) != nameLen {
			continue
		}
		match := true
		for j := 0; j < nameLen; j++ {
			if e.name[j] != name[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}

func copyName(dst *fileEntry, src *[maxName]byte, nameLen int) {
	dst.nameLen = uint8(nameLen)
	for i := 0; i < maxName; i++ {
		dst.name[i] = 0
	}
	for i := 0; i < nameLen; i++ {
		dst.name[i] = src[i]
	}
}
