package tss

import "testing"

func TestSetRSP0WritesLittleEndianAtExpectedOffset(t *testing.T) {
	var tss [TSSSize]byte
	const want uint64 = 0x1122334455667788

	SetRSP0(&tss, want)

	for i := 0; i < 8; i++ {
		got := tss[tssRSP0Offset+i]
		exp := byte(want >> (8 * i))
		if got != exp {
			t.Fatalf("RSP0 byte[%d] mismatch: got=0x%02x want=0x%02x", i, got, exp)
		}
	}
}

func TestSetIomapBaseWritesLittleEndianAtExpectedOffset(t *testing.T) {
	var tss [TSSSize]byte
	const want uint16 = 0xABCD

	SetIomapBase(&tss, want)

	if got := uint16(tss[tssIomapBaseOffset]) | uint16(tss[tssIomapBaseOffset+1])<<8; got != want {
		t.Fatalf("iomap base mismatch: got=0x%04x want=0x%04x", got, want)
	}
}

func TestEncodeTSSDescriptorEncodesBaseAndLimit(t *testing.T) {
	const (
		base  = uintptr(0x1122334455667788)
		limit = uint32(0x00000067)
	)

	low, high := EncodeTSSDescriptor(base, limit)

	if got := uint32(low & 0xFFFF); got != (limit & 0xFFFF) {
		t.Fatalf("limit low mismatch: got=0x%x want=0x%x", got, limit&0xFFFF)
	}

	if got := uint64((low >> 16) & 0xFFFFFF); got != (uint64(base) & 0xFFFFFF) {
		t.Fatalf("base[23:0] mismatch: got=0x%x want=0x%x", got, uint64(base)&0xFFFFFF)
	}

	if got := uint64((low >> 40) & 0xFF); got != 0x89 {
		t.Fatalf("access byte mismatch: got=0x%x want=0x89", got)
	}

	if got := uint64((low >> 48) & 0xF); got != (uint64(limit)>>16)&0xF {
		t.Fatalf("limit high nibble mismatch: got=0x%x want=0x%x", got, (uint64(limit)>>16)&0xF)
	}

	if got := uint64((low >> 56) & 0xFF); got != (uint64(base)>>24)&0xFF {
		t.Fatalf("base[31:24] mismatch: got=0x%x want=0x%x", got, (uint64(base)>>24)&0xFF)
	}

	if got := high; got != uint64(base)>>32 {
		t.Fatalf("base high mismatch: got=0x%x want=0x%x", got, uint64(base)>>32)
	}
}
