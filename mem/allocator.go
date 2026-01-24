package mem

import "unsafe"

const pageSize = 4096

func bootstrapEnd() uint64
func kernelEnd() uint64

var (
	pfaReady    bool
	totalPages  uint64
	freePages   uint64
	bitmapPhys  uint64 // Physical address of the bitmap
	bitmapBytes uint64
	scanStart   uint64 // First page index to start scanning from
)

func kernelEndPhys() uint64 {
	kend := kernelEnd()
	bend := bootstrapEnd()
	if bend > kend {
		return bend
	}
	return kend
}

func alignUp(v, a uint64) uint64 {
	return (v + (a - 1)) & ^(a - 1)
}

func alignDown(v, a uint64) uint64 {
	return v & ^(a - 1)
}

func u64FromHiLo(hi, lo uint32) uint64 {
	return (uint64(hi) << 32) | uint64(lo)
}

func maxAvailableEnd() uint64 {
	// returns the highest end address among all available (type=1) regions
	var maxEnd uint64
	for i := 0; i < mmapCount; i++ {
		e := mmapEntries[i]
		if e.typ != 1 {
			continue
		}
		base := u64FromHiLo(e.baseHi, e.baseLo)
		l := u64FromHiLo(e.lenHi, e.lenLo)
		end := base + l
		if end > maxEnd {
			maxEnd = end
		}
	}
	return maxEnd
}

func bitmapBytePtr(off uint64) *byte {
	// assumes identity mapping / paging off: physical == directly addressable pointer
	return (*byte)(unsafe.Pointer(uintptr(bitmapPhys) + uintptr(off)))
}

func bitmapGet(page uint64) bool {
	byteIdx := page >> 3
	bit := byte(1 << (page & 7))
	b := *bitmapBytePtr(byteIdx)
	return (b & bit) != 0
}

func bitmapSet(page uint64, used bool) {
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

func markFreeRange(startPhys, endPhys uint64) {
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

func markUsedRange(startPhys, endPhys uint64) {
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
	totalPages = (maxEnd + (pageSize - 1)) / pageSize
	if totalPages == 0 {
		return false
	}

	bitmapBytes = (totalPages + 7) / 8

	// place bitmap inside a usable memory region
	kend := kernelEndPhys()
	bitmapPhys = 0
	for i := 0; i < mmapCount; i++ {
		e := mmapEntries[i]
		if e.typ != 1 {
			continue
		}
		base := u64FromHiLo(e.baseHi, e.baseLo)
		end := base + u64FromHiLo(e.lenHi, e.lenLo)

		if end <= base || end <= kend {
			continue
		}

		start := base
		if start < kend {
			start = kend
		}
		start = alignUp(start, pageSize)
		if start+bitmapBytes <= end {
			bitmapPhys = start
			break
		}
	}
	if bitmapPhys == 0 {
		// fallback to just after the kernel end
		bitmapPhys = alignUp(kend, pageSize)
	}

	// reserve full pages for the bitmap
	bitmapPages := alignUp(bitmapBytes, pageSize) / pageSize
	bitmapEnd := bitmapPhys + bitmapPages*pageSize

	// start with everything marked as used
	for i := uint64(0); i < bitmapBytes; i++ {
		*bitmapBytePtr(i) = 0xFF
	}

	// free pages that belong to "available" memory regions (type=1)
	for i := 0; i < mmapCount; i++ {
		e := mmapEntries[i]
		if e.typ != 1 {
			continue
		}

		start := u64FromHiLo(e.baseHi, e.baseLo)
		end := start + u64FromHiLo(e.lenHi, e.lenLo)
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

func TotalPages() uint64 { return totalPages }
func FreePages() uint64  { return freePages }
func UsedPages() uint64  { return totalPages - freePages }

func AllocPage() uint64 {
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

func FreePage(addr uint64) bool {
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
