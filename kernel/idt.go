package kernel

import (
	"unsafe"

	"github.com/dmarro89/go-dav-os/terminal"
)

const (
	idtSize            = 256
	intGateKernelFlags = 0x8E // P=1, DPL=0, interrupt gate
	intGateUserFlags   = 0xEE // P=1, DPL=3, interrupt gate (syscall)
)

const (
	SYS_WRITE = 1
	// SYS_EXIT  = 2 // Not implemented
)

type TrapFrame struct {
	R15    uint64
	R14    uint64
	R13    uint64
	R12    uint64
	R11    uint64
	R10    uint64
	R9     uint64
	R8     uint64
	RDI    uint64
	RSI    uint64
	RBP    uint64
	RBX    uint64
	RDX    uint64
	RCX    uint64
	RAX    uint64
	RIP    uint64
	CS     uint64
	RFLAGS uint64
}

// 16 bytes (x86_64 IDT entry)
type idtEntry struct {
	offsetLow  uint16
	selector   uint16
	ist        uint8
	flags      uint8
	offsetMid  uint16
	offsetHigh uint32
	zero       uint32
}

var idt [idtSize]idtEntry
var idtr [10]byte

// Assembly hooks (boot/stubs_amd64.s)
func LoadIDT(p *[10]byte)
func StoreIDT(p *[10]byte)

func getInt80StubAddr() uint64
func getGPFaultStubAddr() uint64
func getDFaultStubAddr() uint64
func Int80Stub()
func TriggerInt80()
func GetCS() uint16
func getIRQ0StubAddr() uint64
func getIRQ1StubAddr() uint64

// syscalls
func TriggerSysWrite(buf *byte, n uint32)

func Int80Handler(tf *TrapFrame) {
	switch uint32(tf.RAX) {
	case SYS_WRITE:
		fd := tf.RBX
		buf := uintptr(tf.RCX)
		n := tf.RDX
		tf.RAX = sysWrite(fd, buf, n)
	default:
		terminal.Print("unknown syscall\n")
		tf.RAX = ^uint64(0) // return -1
	}
}

func sysWrite(fd uint64, buf uintptr, n uint64) uint64 {
	if fd != 1 {
		return ^uint64(0)
	}

	for i := uint64(0); i < n; i++ {
		b := *(*byte)(unsafe.Pointer(buf + uintptr(i)))
		terminal.PutRune(rune(b))
	}
	return n
}

func packIDTR(limit uint16, base uint64, out *[10]byte) {
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

// func unpackIDTR(in *[6]byte) (limit uint16, base uint32) {
// 	limit = uint16(in[0]) | uint16(in[1])<<8
// 	base = uint32(in[2]) |
// 		uint32(in[3])<<8 |
// 		uint32(in[4])<<16 |
// 		uint32(in[5])<<24
// 	return
// }

func setIDTEntry(vec uint8, handler uint64, selector uint16, flags uint8) {
	e := &idt[vec]
	e.offsetLow = uint16(handler & 0xFFFF)
	e.selector = selector
	e.ist = 0
	e.flags = flags
	e.offsetMid = uint16((handler >> 16) & 0xFFFF)
	e.offsetHigh = uint32((handler >> 32) & 0xFFFFFFFF)
	e.zero = 0
}

// InitIDT builds the IDT and loads it into the CPU
func InitIDT() {
	cs := GetCS()

	// Install emergency handlers first
	setIDTEntry(0x08, getDFaultStubAddr(), cs, intGateKernelFlags)  // #DF
	setIDTEntry(0x0D, getGPFaultStubAddr(), cs, intGateKernelFlags) // #GP

	// Install IRQ handlers
	setIDTEntry(0x20, getIRQ0StubAddr(), cs, intGateKernelFlags) // IRQ0
	setIDTEntry(0x21, getIRQ1StubAddr(), cs, intGateKernelFlags) // IRQ1

	// Install 0x80 syscall handler
	setIDTEntry(0x80, getInt80StubAddr(), cs, intGateUserFlags)

	// Build IDTR (packed 10 bytes)
	base := uint64(uintptr(unsafe.Pointer(&idt[0])))
	limit := uint16(idtSize*16 - 1)
	packIDTR(limit, base, &idtr)

	LoadIDT(&idtr)

	// For testing purposes, read back from CPU (sidt) and print the results
	// storedLimit, storedBase := readIDTR()
	// terminal.Print("IDT limit=")
	// printHex16(storedLimit)
	// terminal.Print(" base=")
	// printHex32(storedBase)
	// terminal.Print("\n")
}

// func readIDTR() (limit uint16, base uint32) {
// 	StoreIDT(&idtr)
// 	return unpackIDTR(&idtr)
// }

// func DumpIDTEntryHW(vec uint8) {
// _, base := readIDTR()

// 	addr := uintptr(base) + uintptr(vec)*8
// 	p := (*[8]byte)(unsafe.Pointer(addr))

// 	terminal.Print("IDT[0x")
// 	printHex8(vec)
// 	terminal.Print("] = ")
// 	for i := 0; i < 8; i++ {
// 		printHex8(p[i])
// 	}
// 	terminal.Print("\n")
// }
