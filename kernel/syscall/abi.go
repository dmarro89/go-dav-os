package syscall

const (
	SysWrite    = 1
	SysExit     = 2
	SysGetTicks = 3
)

type TrapFrame struct {
	R15       uint64
	R14       uint64
	R13       uint64
	R12       uint64
	R11       uint64
	R10       uint64
	R9        uint64
	R8        uint64
	RDI       uint64
	RSI       uint64
	RBP       uint64
	RBX       uint64
	RDX       uint64
	RCX       uint64
	RAX       uint64
	ErrorCode uint64
	RIP       uint64
	CS        uint64
	RFLAGS    uint64
	RSP       uint64
	SS        uint64
}
