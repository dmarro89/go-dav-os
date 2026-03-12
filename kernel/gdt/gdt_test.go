package gdt

import "testing"

func TestPackGDTRPacksLimitAndBase(t *testing.T) {
	var gdtr [10]byte
	const (
		limit = uint16(0xABCD)
		base  = uint64(0x1122334455667788)
	)

	PackGDTR(limit, base, &gdtr)

	gotLimit := uint16(gdtr[0]) | uint16(gdtr[1])<<8
	if gotLimit != limit {
		t.Fatalf("limit mismatch: got=0x%04x want=0x%04x", gotLimit, limit)
	}

	gotBase := uint64(gdtr[2]) |
		uint64(gdtr[3])<<8 |
		uint64(gdtr[4])<<16 |
		uint64(gdtr[5])<<24 |
		uint64(gdtr[6])<<32 |
		uint64(gdtr[7])<<40 |
		uint64(gdtr[8])<<48 |
		uint64(gdtr[9])<<56
	if gotBase != base {
		t.Fatalf("base mismatch: got=0x%x want=0x%x", gotBase, base)
	}
}
