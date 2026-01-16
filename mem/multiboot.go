package mem

import "unsafe"

const maxMMapEntries = 64

// mmapEntry is a normalized copy of one Multiboot2 memory map entry
type mmapEntry struct {
	baseLo uint32
	baseHi uint32
	lenLo  uint32
	lenHi  uint32
	typ    uint32
}

var (
	// mmapEntries stores a compact snapshot of the memory map provided by GRUB
	mmapEntries [maxMMapEntries]mmapEntry
	mmapCount   int
)

const (
	multiboot2TagTypeEnd  = 0
	multiboot2TagTypeMmap = 6
)

// readU32 reads a 32-bit value from memory at the given address
func readU32(addr uintptr) uint32 {
	return *(*uint32)(unsafe.Pointer(addr))
}

// readU64 reads a 64-bit value from memory at the given address
func readU64(addr uintptr) uint64 {
	return *(*uint64)(unsafe.Pointer(addr))
}

func alignUp8(p uintptr) uintptr {
	return (p + 7) &^ 7
}

// InitMultiboot initializes the memory map from the Multiboot info structure
// Returns true if the memory map is valid, false otherwise
func InitMultiboot(mbInfoAddr uint64) bool {
	// reset the memory map counter
	mmapCount = 0
	if mbInfoAddr == 0 {
		return false
	}

	info := uintptr(mbInfoAddr)
	totalSize := readU32(info)
	if totalSize < 16 {
		return false
	}

	foundMmap := false
	p := info + 8
	end := info + uintptr(totalSize)

	for p+8 <= end {
		tagType := readU32(p)
		tagSize := readU32(p + 4)

		if tagType == multiboot2TagTypeEnd {
			break
		}
		if tagSize < 8 || p+uintptr(tagSize) > end {
			break
		}

		if tagType == multiboot2TagTypeMmap {
			if tagSize >= 16 {
				entrySize := readU32(p + 8)
				_ = readU32(p + 12) // entry_version
				if entrySize >= 24 {
					entriesStart := p + 16
					entriesEnd := p + uintptr(tagSize)
					for ep := entriesStart; ep+uintptr(entrySize) <= entriesEnd && mmapCount < maxMMapEntries; ep += uintptr(entrySize) {
						base := readU64(ep)
						length := readU64(ep + 8)
						typ := readU32(ep + 16)

						mmapEntries[mmapCount] = mmapEntry{
							baseLo: uint32(base), baseHi: uint32(base >> 32),
							lenLo: uint32(length), lenHi: uint32(length >> 32),
							typ: typ,
						}
						mmapCount++
					}
					foundMmap = mmapCount > 0
				}
			}
		}

		p = alignUp8(p + uintptr(tagSize))
	}

	return foundMmap
}

// MMapCount returns the number of memory map entries
func MMapCount() int { return mmapCount }

// MMapEntry returns the memory map entry at the given index
func MMapEntry(i int) (baseLo, baseHi, lenLo, lenHi, typ uint32) {
	if i < 0 || i >= mmapCount {
		return 0, 0, 0, 0, 0
	}
	e := mmapEntries[i]
	return e.baseLo, e.baseHi, e.lenLo, e.lenHi, e.typ
}
