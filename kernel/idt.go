//go:build !testing

package kernel

import (
	"unsafe"

	"github.com/dmarro89/go-dav-os/kernel/syscall"
	"github.com/dmarro89/go-dav-os/terminal"
)

const (
	idtSize            = 256
	intGateKernelFlags = 0x8E // P=1, DPL=0, interrupt gate
	intGateUserFlags   = 0xEE // P=1, DPL=3, interrupt gate (syscall)
)

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
func getPFaultStubAddr() uint64
func getDFaultStubAddr() uint64
func Int80Stub()
func TriggerInt80()
func GetCS() uint16
func GetCR2() uint64
func getIRQ0StubAddr() uint64
func getIRQ1StubAddr() uint64

func GPFaultHandler(tf *syscall.TrapFrame) {
	if tf.CS&3 == 3 {
		terminal.Print("\n#GP in user mode\n")
		printFaultDiagnostics("General Protection Fault", tf)
		ReturnToKernel()
	} else {
		terminal.Print("\n#GP in kernel mode\n")
		printFaultDiagnostics("General Protection Fault", tf)
		for {
		} // Halt
	}
}

func PFaultHandler(tf *syscall.TrapFrame) {
	cr2 := GetCR2()
	if tf.CS&3 == 3 {
		terminal.Print("\n#PF in user mode\n")
		printFaultDiagnostics("Page Fault", tf)
		terminal.Print("CR2: ")
		terminal.PrintHex(cr2)
		terminal.Print("\n")
		ReturnToKernel()
	} else {
		terminal.Print("\n#PF in kernel mode\n")
		printFaultDiagnostics("Page Fault", tf)
		terminal.Print("CR2: ")
		terminal.PrintHex(cr2)
		terminal.Print("\n")
		for {
		} // Halt
	}
}

func printFaultDiagnostics(name string, tf *syscall.TrapFrame) {
	terminal.Print("Fault: ")
	terminal.Print(name)
	terminal.Print("\nRIP: ")
	terminal.PrintHex(tf.RIP)
	terminal.Print("\nError Code: ")
	terminal.PrintHex(tf.ErrorCode)
	terminal.Print("\n")
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
	setIDTEntry(0x0E, getPFaultStubAddr(), cs, intGateKernelFlags)  // #PF

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
}
