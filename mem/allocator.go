package mem

import "unsafe"

const pageSize = 4096

// __kernel_end is the first free physical byte after the kernel
var __kernel_end byte

var (
	pfaReady    bool
	totalPages  uint32
	freePages   uint32
	bitmapPhys  uint32 // Physical address of the bitmap
	bitmapBytes uint32
	scanStart   uint32 // First page index to start scanning from
)

func kernelEndPhys() uint32 {
	return uint32(uintptr(unsafe.Pointer(&__kernel_end)))
}

func alignUp(v, a uint32) uint32 {
	return (v + (a - 1)) & ^(a - 1)
}

func alignDown(v, a uint32) uint32 {
	return v & ^(a - 1)
}

func u64FromHiLo(hi, lo uint32) uint64 {
	return (uint64(hi) << 32) | uint64(lo)
}

func maxAvailableEnd() uint64 {
	// returns the highest end address among all available (type=1) regions
	var max uint64
	for i := 0; i < mmapCount; i++ {
		e := mmapEntries[i]
		if e.typ != 1 {
			continue
		}
		base := u64FromHiLo(e.baseHi, e.baseLo)
		l := u64FromHiLo(e.lenHi, e.lenLo)
		end := base + l
		if end > max {
			max = end
		}
	}
	return max
}

func bitmapBytePtr(off uint32) *byte {
	// assumes identity mapping / paging off: physical == directly addressable pointer
	return (*byte)(unsafe.Pointer(uintptr(bitmapPhys) + uintptr(off)))
}

func bitmapGet(page uint32) bool {
	byteIdx := page >> 3
	bit := byte(1 << (page & 7))
	b := *bitmapBytePtr(byteIdx)
	return (b & bit) != 0
}

func bitmapSet(page uint32, used bool) {
	byteIdx := page >> 3
	bit := byte(1 << (page & 7))
	p := bitmapBytePtr(byteIdx)
	b := *p
	if used {
		*p = b | bit
	} else {
		*p = b &^ bit
	}
}

func markFreeRange(startPhys, endPhys uint32) {
	// marks pages as free inside [startPhys, endPhys)
	if endPhys <= startPhys {
		return
	}

	start := alignUp(startPhys, pageSize)
	end := alignDown(endPhys, pageSize)

	for addr := start; addr < end; addr += pageSize {
		page := addr / pageSize
		if page >= totalPages {
			break
		}
		if bitmapGet(page) {
			bitmapSet(page, false)
			freePages++
		}
	}
}

func markUsedRange(startPhys, endPhys uint32) {
	// marks pages as used inside [startPhys, endPhys)
	if endPhys <= startPhys {
		return
	}

	start := alignDown(startPhys, pageSize)
	end := alignUp(endPhys, pageSize)

	for addr := start; addr < end; addr += pageSize {
		page := addr / pageSize
		if page >= totalPages {
			break
		}
		if !bitmapGet(page) {
			bitmapSet(page, true)
			if freePages > 0 {
				freePages--
			}
		}
	}
}

func InitPFA() bool {
	pfaReady = false
	freePages = 0

	maxEnd := maxAvailableEnd()
	if maxEnd == 0 {
		return false
	}

	// manage up to the highest "available" end address
	totalPages = uint32((maxEnd + (pageSize - 1)) / pageSize)
	if totalPages == 0 {
		return false
	}

	bitmapBytes = (totalPages + 7) / 8

	// place bitmap right after the kernel end, aligned to 4KB
	kend := kernelEndPhys()
	bitmapPhys = alignUp(kend, pageSize)

	// reserve full pages for the bitmap
	bitmapPages := alignUp(bitmapBytes, pageSize) / pageSize
	bitmapEnd := bitmapPhys + bitmapPages*pageSize

	// start with everything marked as used
	for i := uint32(0); i < bitmapBytes; i++ {
		*bitmapBytePtr(i) = 0xFF
	}

	// free pages that belong to "available" memory regions (type=1)
	for i := 0; i < mmapCount; i++ {
		e := mmapEntries[i]
		if e.typ != 1 {
			continue
		}

		// for now, only support regions below 4GiB
		if e.baseHi != 0 || e.lenHi != 0 {
			continue
		}

		start := e.baseLo
		end := e.baseLo + e.lenLo
		markFreeRange(start, end)
	}

	// reserve low memory + kernel + bitmap pages
	// this ensures we never allocate pages overlapping our own data structures
	markUsedRange(0, bitmapEnd)

	// prefer scanning from the end of our reserved area
	scanStart = bitmapEnd / pageSize

	pfaReady = true
	return true
}

func PFAReady() bool { return pfaReady }

func TotalPages() uint32 { return totalPages }
func FreePages() uint32  { return freePages }
func UsedPages() uint32  { return totalPages - freePages }

func AllocPage() uint32 {
	// returns a physical address of a 4KB page, or 0 on failure
	if !pfaReady {
		return 0
	}

	for page := scanStart; page < totalPages; page++ {
		if !bitmapGet(page) {
			bitmapSet(page, true)
			if freePages > 0 {
				freePages--
			}
			return page * pageSize
		}
	}

	return 0
}

func FreePage(addr uint32) bool {
	// frees a page previously returned by AllocPage
	if !pfaReady {
		return false
	}
	if (addr % pageSize) != 0 {
		return false
	}

	page := addr / pageSize
	if page >= totalPages {
		return false
	}
	if !bitmapGet(page) {
		return false
	}

	bitmapSet(page, false)
	freePages++
	return true
}
