package gdt

func PackGDTR(limit uint16, base uint64, out *[10]byte) {
	out[0] = byte(limit)
	out[1] = byte(limit >> 8)
	out[2] = byte(base)
	out[3] = byte(base >> 8)
	out[4] = byte(base >> 16)
	out[5] = byte(base >> 24)
	out[6] = byte(base >> 32)
	out[7] = byte(base >> 40)
	out[8] = byte(base >> 48)
	out[9] = byte(base >> 56)
}
