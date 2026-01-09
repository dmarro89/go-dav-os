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
	EDI    uint32
	ESI    uint32
	EBP    uint32
	ESP    uint32
	EBX    uint32
	EDX    uint32
	ECX    uint32
	EAX    uint32
	EIP    uint32
	CS     uint32
	EFLAGS uint32
}

// 8 byte
type idtEntry struct {
	offsetLow  uint16
	selector   uint16
	zero       uint8
	flags      uint8
	offsetHigh uint16
}

var idt [idtSize]idtEntry
var idtr [6]byte

// Assembly hooks (boot.s)
func LoadIDT(p *[6]byte)
func StoreIDT(p *[6]byte)

func getInt80StubAddr() uint32
func getGPFaultStubAddr() uint32
func getDFaultStubAddr() uint32
func Int80Stub()
func TriggerInt80()
func GetCS() uint16
func getIRQ0StubAddr() uint32
func getIRQ1StubAddr() uint32

// syscalls
func TriggerSysWrite(buf *byte, n uint32)

func Int80Handler(tf *TrapFrame) {
	switch tf.EAX {
	case SYS_WRITE:
		fd := tf.EBX
		buf := tf.ECX
		n := tf.EDX
		tf.EAX = sysWrite(fd, buf, n)
	default:
		terminal.Print("unknown syscall\n")
		tf.EAX = ^uint32(0) // return -1
	}
}

func sysWrite(fd, buf, n uint32) uint32 {
	if fd != 1 || n == 0 {
		return 0
	}

	p := (*[1 << 20]byte)(unsafe.Pointer(uintptr(buf)))[:n:n]

	terminal.Print(string(p))
	return n
}

func packIDTR(limit uint16, base uint32, out *[6]byte) {
	out[0] = byte(limit)
	out[1] = byte(limit >> 8)
	out[2] = byte(base)
	out[3] = byte(base >> 8)
	out[4] = byte(base >> 16)
	out[5] = byte(base >> 24)
}

// func unpackIDTR(in *[6]byte) (limit uint16, base uint32) {
// 	limit = uint16(in[0]) | uint16(in[1])<<8
// 	base = uint32(in[2]) |
// 		uint32(in[3])<<8 |
// 		uint32(in[4])<<16 |
// 		uint32(in[5])<<24
// 	return
// }

func setIDTEntry(vec uint8, handler uint32, selector uint16, flags uint8) {
	e := &idt[vec]
	e.offsetLow = uint16(handler & 0xFFFF)
	e.selector = selector
	e.zero = 0
	e.flags = flags
	e.offsetHigh = uint16((handler >> 16) & 0xFFFF)
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

	// Build IDTR (packed 6 bytes)
	base := uint32(uintptr(unsafe.Pointer(&idt[0])))
	limit := uint16(idtSize*8 - 1)
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
