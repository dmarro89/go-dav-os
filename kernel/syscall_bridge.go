//go:build !testing

package kernel

import ksyscall "github.com/dmarro89/go-dav-os/kernel/syscall"

func TriggerSysWrite(buf *byte, n uint32)
func TriggerSysExit(status uint32)
func TriggerSysGetTicks() uint64
func ReturnToKernel()
func ReadMSR(msr uint32) uint64
func WriteMSR(msr uint32, value uint64)
func getSyscallEntryAddr() uint64

func InitSyscall() {
	ksyscall.Init(ReadMSR, WriteMSR, getSyscallEntryAddr(), kernelCodeSelector, userCodeSelector)
}

func Int80Handler(tf *ksyscall.TrapFrame) {
	ksyscall.Dispatch(tf, GetTicks, ReturnToKernel)
}

func SyscallHandler(tf *ksyscall.TrapFrame) {
	ksyscall.Dispatch(tf, GetTicks, ReturnToKernel)
}
