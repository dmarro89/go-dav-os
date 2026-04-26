//go:build testing

package kernel

func DebugChar(c byte) {}

func inb(port uint16) byte {
	return 0
}

func outb(port uint16, val byte) {}

func EnableInterrupts() {}

func DisableInterrupts() {}

func Halt() {}

func LoadGDT(p *[10]byte) {}

func LoadDataSegments(sel uint16) {}

func LoadIDT(p *[10]byte) {}

func StoreIDT(p *[10]byte) {}

func getInt80StubAddr() uint64 {
	return 0
}

func getGPFaultStubAddr() uint64 {
	return 0
}

func getDFaultStubAddr() uint64 {
	return 0
}

func getPFaultStubAddr() uint64 {
	return 0
}

func Int80Stub() {}

func TriggerInt80() {}

func GetCS() uint16 {
	return 0
}

func getIRQ0StubAddr() uint64 {
	return 0
}

func getIRQ1StubAddr() uint64 {
	return 0
}

func TriggerSysWrite(buf *byte, n uint32) {}

func TriggerSysExit(status uint32) {}

func TriggerSysGetTicks() uint64 {
	return 0
}

func ReturnToKernel() {}

func LoadTR(sel uint16) {}

func ExecuteUserTask(rip, rsp uint64) {}

func GetUserProgramHelloAddr() uint64 {
	return 0
}

func GetUserProgramKernelReadProbeAddr() uint64 {
	return 0
}

func GetUserProgramKernelWriteProbeAddr() uint64 {
	return 0
}

func GetUserStackTopAddr() uint64 {
	return 0
}
