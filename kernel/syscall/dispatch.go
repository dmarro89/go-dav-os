package syscall

import (
	"unsafe"

	"github.com/dmarro89/go-dav-os/terminal"
)

const (
	userVAStart      uintptr = 0x40000000
	userVAEnd        uintptr = 0x40002000
	maxSysWriteBytes         = 4096
	syscallError             = ^uint64(0)
)

var sysWriteBuffer [maxSysWriteBytes]byte

func Dispatch(tf *TrapFrame, getTicks func() uint64, returnToKernel func()) {
	switch uint32(tf.RAX) {
	case SysWrite:
		fd := tf.RDI
		buf := uintptr(tf.RSI)
		n := tf.RDX
		tf.RAX = sysWrite(fd, buf, n)
	case SysExit:
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
	case SysGetTicks:
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
	return sysWriteWithCopier(fd, buf, n, copyFromUserBytes)
}

func sysWriteWithCopier(fd uint64, buf uintptr, n uint64, copier func(*[maxSysWriteBytes]byte, int, uintptr) bool) uint64 {
	if fd != 1 {
		return syscallError
	}
	if n == 0 {
		return 0
	}
	if n > maxSysWriteBytes {
		n = maxSysWriteBytes
	}

	count := int(n)
	if !copier(&sysWriteBuffer, count, buf) {
		return syscallError
	}

	for i := 0; i < count; i++ {
		b := sysWriteBuffer[i]
		terminal.PutRune(rune(b))
	}
	return n
}

func copyFromUserBytes(dst *[maxSysWriteBytes]byte, count int, userPtr uintptr) bool {
	if count < 0 || count > maxSysWriteBytes {
		return false
	}
	if !validUserRange(userPtr, uintptr(count)) {
		return false
	}

	for i := 0; i < count; i++ {
		dst[i] = *(*byte)(unsafe.Pointer(userPtr + uintptr(i)))
	}
	return true
}

func validUserRange(start, length uintptr) bool {
	if length == 0 {
		return true
	}
	if start < userVAStart || start >= userVAEnd {
		return false
	}
	return length <= userVAEnd-userVAStart && start <= userVAEnd-length
}
