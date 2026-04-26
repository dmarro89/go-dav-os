//go:build testing

package shell

func SetGetTicks(fn func() uint64) {}

func SetGetSyscallTicks(fn func() uint64) {}

func SetRunProgram(fn func(name *[16]byte, nameLen int) (pid int, ok bool)) {}

func Run() {}
