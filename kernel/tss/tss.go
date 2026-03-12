package tss

const (
	TSSSize = 104

	tssRSP0Offset      = 4
	tssIomapBaseOffset = 102
)

func SetRSP0(tss *[TSSSize]byte, rsp0 uint64) {
	put64(tss, tssRSP0Offset, rsp0)
}

func SetIomapBase(tss *[TSSSize]byte, iomapBase uint16) {
	put16(tss, tssIomapBaseOffset, iomapBase)
}

func EncodeTSSDescriptor(base uintptr, limit uint32) (low uint64, high uint64) {
	base64 := uint64(base)
	limit64 := uint64(limit)

	// 64-bit available TSS descriptor (type=0x9, present=1).
	low = (limit64 & 0xFFFF) |
		((base64 & 0xFFFFFF) << 16) |
		(uint64(0x89) << 40) |
		(((limit64 >> 16) & 0xF) << 48) |
		(((base64 >> 24) & 0xFF) << 56)
	high = base64 >> 32
	return
}

func put16(dst *[TSSSize]byte, off int, value uint16) {
	dst[off+0] = byte(value)
	dst[off+1] = byte(value >> 8)
}

func put64(dst *[TSSSize]byte, off int, value uint64) {
	dst[off+0] = byte(value)
	dst[off+1] = byte(value >> 8)
	dst[off+2] = byte(value >> 16)
	dst[off+3] = byte(value >> 24)
	dst[off+4] = byte(value >> 32)
	dst[off+5] = byte(value >> 40)
	dst[off+6] = byte(value >> 48)
	dst[off+7] = byte(value >> 56)
}
