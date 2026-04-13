package syscall

import (
	"testing"
	"unsafe"
)

func TestSTARValueUsesKernelAndSYSRETBaseSelectors(t *testing.T) {
	const kernelCS = uint16(0x08)
	const userCS = uint16(0x1B)

	got := STARValue(kernelCS, userCS)
	wantUserBase := (userCS - 16) &^ uint16(3)
	want := uint64(kernelCS&^3)<<32 | uint64(wantUserBase)<<48
	if got != want {
		t.Fatalf("STAR mismatch: got=0x%016x want=0x%016x", got, want)
	}
}

func TestSFMASKValueClearsInterruptFlag(t *testing.T) {
	if got, want := SFMASKValue(), uint64(1)<<9; got != want {
		t.Fatalf("SFMASK mismatch: got=0x%016x want=0x%016x", got, want)
	}
}

func TestEnableSCESetsBitZero(t *testing.T) {
	if got := EnableSCE(0); got != 1 {
		t.Fatalf("EnableSCE mismatch: got=0x%016x want=0x0000000000000001", got)
	}
}

func TestDispatchWriteUsesSyscallABIRegisters(t *testing.T) {
	buf := []byte("test")
	tf := TrapFrame{
		RAX: SysWrite,
		RDI: 1,
		RSI: uint64(uintptr(unsafe.Pointer(&buf[0]))),
		RDX: uint64(len(buf)),
		RBX: 99,
		RCX: 88,
	}

	Dispatch(&tf, nil, nil)

	if tf.RAX != uint64(len(buf)) {
		t.Fatalf("sys_write return mismatch: got=%d want=%d", tf.RAX, len(buf))
	}
}

func TestDispatchKernelExitRejected(t *testing.T) {
	tf := TrapFrame{
		RAX: SysExit,
		RDI: 7,
		CS:  0x08,
	}

	Dispatch(&tf, nil, nil)

	if tf.RAX != ^uint64(0) {
		t.Fatalf("kernel SYS_EXIT should fail: got=0x%016x", tf.RAX)
	}
}

func TestDispatchGetTicksUsesHook(t *testing.T) {
	tf := TrapFrame{RAX: SysGetTicks}

	Dispatch(&tf, func() uint64 { return 1234 }, nil)

	if tf.RAX != 1234 {
		t.Fatalf("SYS_GETTICKS mismatch: got=%d want=1234", tf.RAX)
	}
}
