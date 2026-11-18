package mem

import (
	"encoding/binary"
	"testing"
	"unsafe"
)

// buildMultibootBuffer creates a valid multiboot2 information structure in memory.
// It includes the total size header, mmap tag (type 6), and end tag (type 0).
func buildMultibootBuffer(entries []mmapEntry) []byte {
	// Header: 8 bytes (total_size uint32, reserved uint32)
	// Mmap tag:
	//   type: uint32 = 6
	//   size: uint32 = 16 + len(entries)*24
	//   entry_size: uint32 = 24
	//   entry_version: uint32 = 0
	//   entries: len(entries) * 24 bytes
	// End tag:
	//   type: uint32 = 0
	//   size: uint32 = 8
	mmapTagSize := 16 + len(entries)*24
	// Align mmap tag size up to 8 bytes if needed
	paddedMmapTagSize := (mmapTagSize + 7) &^ 7

	endTagSize := 8

	totalSize := 8 + paddedMmapTagSize + endTagSize
	buf := make([]byte, totalSize)

	// Header
	binary.LittleEndian.PutUint32(buf[0:4], uint32(totalSize))
	binary.LittleEndian.PutUint32(buf[4:8], 0) // reserved

	// Mmap tag
	p := 8
	binary.LittleEndian.PutUint32(buf[p:p+4], multiboot2TagTypeMmap)
	binary.LittleEndian.PutUint32(buf[p+4:p+8], uint32(mmapTagSize))
	binary.LittleEndian.PutUint32(buf[p+8:p+12], 24) // entry_size
	binary.LittleEndian.PutUint32(buf[p+12:p+16], 0) // entry_version

	ep := p + 16
	for _, e := range entries {
		base := uint64(e.baseLo) | (uint64(e.baseHi) << 32)
		length := uint64(e.lenLo) | (uint64(e.lenHi) << 32)
		binary.LittleEndian.PutUint64(buf[ep:ep+8], base)
		binary.LittleEndian.PutUint64(buf[ep+8:ep+16], length)
		binary.LittleEndian.PutUint32(buf[ep+16:ep+20], e.typ)
		binary.LittleEndian.PutUint32(buf[ep+20:ep+24], 0) // reserved
		ep += 24
	}

	// End tag
	pEnd := 8 + paddedMmapTagSize
	binary.LittleEndian.PutUint32(buf[pEnd:pEnd+4], multiboot2TagTypeEnd)
	binary.LittleEndian.PutUint32(buf[pEnd+4:pEnd+8], uint32(endTagSize))

	return buf
}

// setupTestPFA installs a memory buffer for bitmap backing store during unit tests.
func setupTestPFA(t *testing.T) {
	t.Helper()
	// Reset all global allocator state
	pfaReady = false
	totalPages = 0
	freePages = 0
	bitmapPhys = 0
	bitmapBytes = 0
	scanStart = 0
	mmapCount = 0
	testingBitmapBacking = nil

	t.Cleanup(func() {
		testingBitmapBacking = nil
		pfaReady = false
		totalPages = 0
		freePages = 0
		bitmapPhys = 0
		bitmapBytes = 0
		scanStart = 0
		mmapCount = 0
	})
}

func TestInitMultiboot(t *testing.T) {
	t.Run("zero_address", func(t *testing.T) {
		if InitMultiboot(0) {
			t.Fatal("expected InitMultiboot(0) to return false")
		}
		if MMapCount() != 0 {
			t.Fatalf("expected MMapCount() == 0, got %d", MMapCount())
		}
	})

	t.Run("buffer_too_small", func(t *testing.T) {
		buf := make([]byte, 12)
		binary.LittleEndian.PutUint32(buf[0:4], 12)
		addr := uint64(uintptr(unsafe.Pointer(&buf[0])))
		if InitMultiboot(addr) {
			t.Fatal("expected InitMultiboot with totalSize < 16 to return false")
		}
	})

	t.Run("valid_mmap_entries", func(t *testing.T) {
		entries := []mmapEntry{
			{baseLo: 0x00000000, baseHi: 0, lenLo: 0x0009FC00, lenHi: 0, typ: 1},
			{baseLo: 0x0009FC00, baseHi: 0, lenLo: 0x00000400, lenHi: 0, typ: 2},
			{baseLo: 0x00100000, baseHi: 0, lenLo: 0x07EE0000, lenHi: 0, typ: 1},
		}

		buf := buildMultibootBuffer(entries)
		addr := uint64(uintptr(unsafe.Pointer(&buf[0])))

		if !InitMultiboot(addr) {
			t.Fatal("expected InitMultiboot to succeed")
		}

		if MMapCount() != len(entries) {
			t.Fatalf("expected MMapCount() == %d, got %d", len(entries), MMapCount())
		}

		for i, expected := range entries {
			bLo, bHi, lLo, lHi, typ := MMapEntry(i)
			if bLo != expected.baseLo || bHi != expected.baseHi ||
				lLo != expected.lenLo || lHi != expected.lenHi ||
				typ != expected.typ {
				t.Errorf("entry %d mismatch: got base=%x:%x len=%x:%x typ=%d, expected base=%x:%x len=%x:%x typ=%d",
					i, bHi, bLo, lHi, lLo, typ, expected.baseHi, expected.baseLo, expected.lenHi, expected.lenLo, expected.typ)
			}
		}

		// Out of bounds MMapEntry test
		bLo, bHi, lLo, lHi, typ := MMapEntry(-1)
		if bLo != 0 || bHi != 0 || lLo != 0 || lHi != 0 || typ != 0 {
			t.Error("expected 0 for negative index")
		}
		bLo, bHi, lLo, lHi, typ = MMapEntry(len(entries))
		if bLo != 0 || bHi != 0 || lLo != 0 || lHi != 0 || typ != 0 {
			t.Error("expected 0 for out of bounds index")
		}
	})

	t.Run("no_mmap_tag_present", func(t *testing.T) {
		// Buffer with only header and end tag
		buf := make([]byte, 16)
		binary.LittleEndian.PutUint32(buf[0:4], 16)
		binary.LittleEndian.PutUint32(buf[8:12], multiboot2TagTypeEnd)
		binary.LittleEndian.PutUint32(buf[12:16], 8)

		addr := uint64(uintptr(unsafe.Pointer(&buf[0])))
		if InitMultiboot(addr) {
			t.Fatal("expected InitMultiboot without mmap tag to return false")
		}
		if MMapCount() != 0 {
			t.Fatalf("expected MMapCount() == 0, got %d", MMapCount())
		}
	})
}

func TestPageFrameAllocator_Uninitialized(t *testing.T) {
	setupTestPFA(t)

	if PFAReady() {
		t.Fatal("expected PFAReady() to be false initially")
	}
	if AllocPage() != 0 {
		t.Fatal("expected AllocPage() on uninitialized PFA to return 0")
	}
	if FreePage(0x1000) {
		t.Fatal("expected FreePage() on uninitialized PFA to return false")
	}
}

func TestPageFrameAllocator_InitWithoutMemory(t *testing.T) {
	setupTestPFA(t)

	// No mmap entries added
	if InitPFA() {
		t.Fatal("expected InitPFA() with 0 available memory to return false")
	}
	if PFAReady() {
		t.Fatal("expected PFAReady() to be false")
	}
}

func TestPageFrameAllocator_AllocAndFree(t *testing.T) {
	setupTestPFA(t)

	// Configure a test memory map:
	// Region 1: 0x00000000 - 0x000A0000 (available, 160 pages)
	// Region 2: 0x00100000 - 0x00120000 (available, 32 pages, from 1MB to 1MB+128KB)
	entries := []mmapEntry{
		{baseLo: 0x00000000, baseHi: 0, lenLo: 0x000A0000, lenHi: 0, typ: 1},
		{baseLo: 0x00100000, baseHi: 0, lenLo: 0x00020000, lenHi: 0, typ: 1},
	}
	buf := buildMultibootBuffer(entries)
	addr := uint64(uintptr(unsafe.Pointer(&buf[0])))
	if !InitMultiboot(addr) {
		t.Fatal("InitMultiboot failed")
	}

	if !InitPFA() {
		t.Fatal("InitPFA failed")
	}

	if !PFAReady() {
		t.Fatal("PFAReady() expected true")
	}

	total := TotalPages()
	initialFree := FreePages()
	used := UsedPages()

	if total == 0 {
		t.Fatal("expected TotalPages > 0")
	}
	if initialFree == 0 {
		t.Fatal("expected FreePages > 0")
	}
	if total != initialFree+used {
		t.Fatalf("total (%d) != free (%d) + used (%d)", total, initialFree, used)
	}

	// Allocate a page
	page1 := AllocPage()
	if page1 == 0 {
		t.Fatal("expected AllocPage to return non-zero page address")
	}
	if page1%pageSize != 0 {
		t.Fatalf("allocated page address 0x%x is not 4KB aligned", page1)
	}
	if FreePages() != initialFree-1 {
		t.Fatalf("expected FreePages to decrement by 1, got %d (was %d)", FreePages(), initialFree)
	}
	if UsedPages() != used+1 {
		t.Fatalf("expected UsedPages to increment by 1, got %d (was %d)", UsedPages(), used)
	}

	// Allocate a second page
	page2 := AllocPage()
	if page2 == 0 {
		t.Fatal("expected second AllocPage to succeed")
	}
	if page2 == page1 {
		t.Fatalf("expected distinct page address, got duplicate 0x%x", page2)
	}
	if page2%pageSize != 0 {
		t.Fatalf("page2 address 0x%x not aligned", page2)
	}

	// Free page1
	if !FreePage(page1) {
		t.Fatalf("expected FreePage(0x%x) to succeed", page1)
	}
	if FreePages() != initialFree-1 {
		t.Fatalf("expected FreePages to be %d, got %d", initialFree-1, FreePages())
	}

	// Double-free should fail
	if FreePage(page1) {
		t.Fatal("expected double-free of page1 to return false")
	}

	// Free page2
	if !FreePage(page2) {
		t.Fatalf("expected FreePage(0x%x) to succeed", page2)
	}
	if FreePages() != initialFree {
		t.Fatalf("expected FreePages to return to initial count %d, got %d", initialFree, FreePages())
	}

	// Allocate again - should reuse freed space
	reallocated := AllocPage()
	if reallocated == 0 {
		t.Fatal("expected AllocPage to succeed after free")
	}
}

func TestPageFrameAllocator_FreeValidation(t *testing.T) {
	setupTestPFA(t)

	entries := []mmapEntry{
		{baseLo: 0x00100000, baseHi: 0, lenLo: 0x00010000, lenHi: 0, typ: 1}, // 64KB (16 pages)
	}
	buf := buildMultibootBuffer(entries)
	addr := uint64(uintptr(unsafe.Pointer(&buf[0])))
	if !InitMultiboot(addr) || !InitPFA() {
		t.Fatal("InitMultiboot or InitPFA failed")
	}

	// Unaligned address
	if FreePage(0x100001) {
		t.Fatal("expected FreePage with unaligned address to return false")
	}

	// Address beyond totalPages * pageSize
	outOfBounds := (TotalPages() + 10) * pageSize
	if FreePage(outOfBounds) {
		t.Fatal("expected FreePage with out-of-bounds address to return false")
	}

	// Address of page that is currently not marked used (already free)
	// Allocate a page, free it, then try freeing again
	p := AllocPage()
	if p == 0 {
		t.Fatal("AllocPage failed")
	}
	if !FreePage(p) {
		t.Fatal("FreePage failed")
	}
	if FreePage(p) {
		t.Fatal("expected FreePage on already-free page to return false")
	}
}

func TestPageFrameAllocator_Exhaustion(t *testing.T) {
	setupTestPFA(t)

	// Small memory map with only 3 usable pages beyond the bitmap
	entries := []mmapEntry{
		{baseLo: 0x00100000, baseHi: 0, lenLo: 0x00006000, lenHi: 0, typ: 1}, // 24KB = 6 pages
	}
	buf := buildMultibootBuffer(entries)
	addr := uint64(uintptr(unsafe.Pointer(&buf[0])))
	if !InitMultiboot(addr) || !InitPFA() {
		t.Fatal("InitMultiboot or InitPFA failed")
	}

	initialFree := FreePages()
	allocated := make([]uint64, 0, initialFree)

	for i := uint64(0); i < initialFree; i++ {
		p := AllocPage()
		if p == 0 {
			t.Fatalf("expected allocation %d to succeed", i)
		}
		allocated = append(allocated, p)
	}

	// Next allocation must fail due to exhaustion
	if AllocPage() != 0 {
		t.Fatal("expected AllocPage to return 0 when exhausted")
	}
	if FreePages() != 0 {
		t.Fatalf("expected FreePages == 0, got %d", FreePages())
	}

	// Free one page
	if !FreePage(allocated[0]) {
		t.Fatal("expected FreePage to succeed")
	}
	if FreePages() != 1 {
		t.Fatalf("expected FreePages == 1, got %d", FreePages())
	}

	// Now allocation should succeed again
	reclaimed := AllocPage()
	if reclaimed != allocated[0] {
		t.Fatalf("expected reclaimed page 0x%x, got 0x%x", allocated[0], reclaimed)
	}

	// Clean up and free all
	for _, p := range allocated[1:] {
		if !FreePage(p) {
			t.Fatalf("failed to free page 0x%x", p)
		}
	}
	if !FreePage(reclaimed) {
		t.Fatalf("failed to free reclaimed page 0x%x", reclaimed)
	}
	if FreePages() != initialFree {
		t.Fatalf("expected FreePages == %d, got %d", initialFree, FreePages())
	}
}
