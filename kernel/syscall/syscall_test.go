package syscall

import "testing"

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
	const userBuf = uintptr(userVAStart)

	copier := func(dst *[maxSysWriteBytes]byte, count int, src uintptr) bool {
		if src != userBuf {
			t.Fatalf("copy_from_user source mismatch: got=0x%x want=0x%x", src, userBuf)
		}
		if count != 4 {
			t.Fatalf("copy_from_user length mismatch: got=%d want=4", count)
		}
		dst[0] = 't'
		dst[1] = 'e'
		dst[2] = 's'
		dst[3] = 't'
		return true
	}

	if got := sysWriteWithCopier(1, userBuf, 4, copier); got != 4 {
		t.Fatalf("sys_write return mismatch: got=%d want=4", got)
	}
}

func TestSysWriteRejectsInvalidUserPointer(t *testing.T) {
	if got := sysWrite(1, 0, 1); got != syscallError {
		t.Fatalf("sys_write invalid pointer mismatch: got=0x%016x want=0x%016x", got, syscallError)
	}
}

func TestSysWriteClampsLargeWrites(t *testing.T) {
	const userBuf = uintptr(userVAStart)

	copier := func(dst *[maxSysWriteBytes]byte, count int, src uintptr) bool {
		if src != userBuf {
			t.Fatalf("copy_from_user source mismatch: got=0x%x want=0x%x", src, userBuf)
		}
		if count != maxSysWriteBytes {
			t.Fatalf("copy_from_user length mismatch: got=%d want=%d", count, maxSysWriteBytes)
		}
		return true
	}

	if got := sysWriteWithCopier(1, userBuf, maxSysWriteBytes+1, copier); got != maxSysWriteBytes {
		t.Fatalf("sys_write clamp mismatch: got=%d want=%d", got, maxSysWriteBytes)
	}
}

func TestValidUserRange(t *testing.T) {
	cases := []struct {
		name   string
		start  uintptr
		length uintptr
		valid  bool
	}{
		{name: "zero length", start: 0, length: 0, valid: true},
		{name: "full user window", start: userVAStart, length: userVAEnd - userVAStart, valid: true},
		{name: "last byte", start: userVAEnd - 1, length: 1, valid: true},
		{name: "before window", start: userVAStart - 1, length: 1, valid: false},
		{name: "at end", start: userVAEnd, length: 1, valid: false},
		{name: "crosses end", start: userVAEnd - 1, length: 2, valid: false},
	}

	for _, tc := range cases {
		if got := validUserRange(tc.start, tc.length); got != tc.valid {
			t.Fatalf("%s: got=%v want=%v", tc.name, got, tc.valid)
		}
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
