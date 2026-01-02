package mem

import "unsafe"

const maxMMapEntries = 64

// mmapEntry is a normalized copy of one Multiboot v1 memory map entry
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

// readU32 reads a 32-bit value from memory at the given address
func readU32(addr uintptr) uint32 {
	return *(*uint32)(unsafe.Pointer(addr))
}

// InitMultiboot initializes the memory map from the Multiboot info structure
// Returns true if the memory map is valid, false otherwise
func InitMultiboot(mbInfoAddr uint32) bool {
	// reset the memory map counter
	mmapCount = 0

	flags := readU32(uintptr(mbInfoAddr) + 0)

	// check if the memory map is valid
	// Multiboot v1: bit 6 -> mmap_* valid
	if (flags & (1 << 6)) == 0 {
		return false
	}

	// get the length and address of the memory map
	// offset 44: mmap_length
	// offset 48: mmap_addr
	mmapLen := readU32(uintptr(mbInfoAddr) + 44)
	mmapAddr := readU32(uintptr(mbInfoAddr) + 48)

	// iterate over the memory map
	p := uintptr(mmapAddr)
	end := p + uintptr(mmapLen)

	for p < end && mmapCount < maxMMapEntries {
		// get the size of the entry
		// offset 0: size
		size := readU32(p)
		// get the base of the entry
		// offset 4: base_lo
		baseLo := readU32(p + 4)
		baseHi := readU32(p + 8)
		// get the length of the entry
		// offset 12: len_lo
		lenLo := readU32(p + 12)
		// get the length of the entry
		// offset 16: len_hi
		lenHi := readU32(p + 16)
		// get the type of the entry
		// offset 20: type
		typ := readU32(p + 20)

		// store the entry in the memory map
		mmapEntries[mmapCount] = mmapEntry{
			baseLo: baseLo, baseHi: baseHi,
			lenLo: lenLo, lenHi: lenHi,
			typ: typ,
		}
		// increment the memory map counter
		mmapCount++

		// increment the pointer by the size of the entry and 4 bytes (padding)
		p += uintptr(size) + 4
	}

	// return true if the memory map is valid
	return true
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
