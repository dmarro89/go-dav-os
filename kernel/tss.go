//go:build !testing

package kernel

import (
	"unsafe"

	tsslib "github.com/dmarro89/go-dav-os/kernel/tss"
)

const kernelTrapStackSize = 4096

var (
	cpuTSS    [tsslib.TSSSize]byte
	trapStack [kernelTrapStackSize]byte
)

func LoadTR(sel uint16)

func InitGDTAndTSS() {
	initGDT()

	tsslib.SetIomapBase(&cpuTSS, tsslib.TSSSize)
	SetKernelRSP0(defaultKernelTrapStackTop())
	setTSSDescriptor(uintptr(unsafe.Pointer(&cpuTSS[0])), tsslib.TSSSize-1)

	loadGDT()
	LoadTR(tssSelector)
}

func SetKernelRSP0(rsp0 uint64) {
	tsslib.SetRSP0(&cpuTSS, rsp0)
}

func defaultKernelTrapStackTop() uint64 {
	top := uintptr(unsafe.Pointer(&trapStack[0])) + uintptr(len(trapStack))
	return uint64(top &^ uintptr(0xF))
}
