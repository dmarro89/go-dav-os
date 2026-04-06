package syscall

import (
	"unsafe"

	"github.com/dmarro89/go-dav-os/terminal"
)

func Dispatch(tf *TrapFrame, getTicks func() uint64, returnToKernel func()) {
	switch uint32(tf.RAX) {
	case SYS_WRITE:
		fd := tf.RDI
		buf := uintptr(tf.RSI)
		n := tf.RDX
		tf.RAX = sysWrite(fd, buf, n)
	case SYS_EXIT:
		status := int(tf.RDI)
		if tf.CS&3 == 3 {
			terminal.Print("Process exited with status ")
			terminal.PrintInt(status)
			terminal.Print("\n")
			if returnToKernel != nil {
				returnToKernel()
			}
			return
		}

		terminal.Print("kernel-mode SYS_EXIT rejected (status ")
		terminal.PrintInt(status)
		terminal.Print(")\n")
		tf.RAX = ^uint64(0)
	case SYS_GETTICKS:
		if getTicks == nil {
			tf.RAX = 0
			return
		}
		tf.RAX = getTicks()
	default:
		terminal.Print("unknown syscall\n")
		tf.RAX = ^uint64(0)
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
