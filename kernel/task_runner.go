package kernel

import "unsafe"

var helloProgramName = [...]byte{'h', 'e', 'l', 'l', 'o'}

const helloProgramNameLen = 5

func ExecuteUserTask(rip, rsp uint64)
func GetUserProgramShellAddr() uint64

var userStack [4096]byte

func RunProgram(name *[16]byte, nameLen int) (pid int, ok bool) {
	if nameLen != helloProgramNameLen {
		return -1, false
	}

	for i := 0; i < helloProgramNameLen; i++ {
		if name[i] != helloProgramName[i] {
			return -1, false
		}
	}

	stackAddr := uint64(uintptr(unsafe.Pointer(&userStack[0])) + uintptr(len(userStack)))
	stackAddr = stackAddr &^ 15

	rip := GetUserProgramShellAddr()
	ExecuteUserTask(rip, stackAddr)

	return 1, true
}
