package kernel

import (
	"unsafe"

	gdtlib "github.com/dmarro89/go-dav-os/kernel/gdt"
	tsslib "github.com/dmarro89/go-dav-os/kernel/tss"
)

const (
	kernelCodeSelector uint16 = 0x08
	kernelDataSelector uint16 = 0x10
	userCodeSelector   uint16 = 0x1B
	userDataSelector   uint16 = 0x23
	tssSelector        uint16 = 0x28
)

const (
	kernelCodeDescriptor uint64 = 0x00AF9A000000FFFF
	kernelDataDescriptor uint64 = 0x00CF92000000FFFF
	userCodeDescriptor   uint64 = 0x00AFFA000000FFFF
	userDataDescriptor   uint64 = 0x00CFF2000000FFFF
)

var (
	gdt         [7]uint64
	gdtRegister [10]byte
)

func LoadGDT(p *[10]byte)
func LoadDataSegments(sel uint16)

func initGDT() {
	gdt[0] = 0
	gdt[1] = kernelCodeDescriptor
	gdt[2] = kernelDataDescriptor
	gdt[3] = userCodeDescriptor
	gdt[4] = userDataDescriptor
}

func setTSSDescriptor(base uintptr, limit uint32) {
	gdt[5], gdt[6] = tsslib.EncodeTSSDescriptor(base, limit)
}

func loadGDT() {
	gdtlib.PackGDTR(uint16(len(gdt)*8-1), uint64(uintptr(unsafe.Pointer(&gdt[0]))), &gdtRegister)
	LoadGDT(&gdtRegister)
	LoadDataSegments(kernelDataSelector)
}
