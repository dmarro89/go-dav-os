package kernel

import "github.com/dmarro89/go-dav-os/kernel/scheduler"

func userHelloStart()

var helloProgramName = [...]byte{'h', 'e', 'l', 'l', 'o'}

const helloProgramNameLen = 5

func RunProgram(name *[16]byte, nameLen int) (pid int, ok bool) {
	if !matchProgramName(name, nameLen) {
		return -1, false
	}
	task := scheduler.NewTask(userHelloStart)
	if task == nil {
		return -1, false
	}
	return task.ID, true
}

func matchProgramName(name *[16]byte, nameLen int) bool {
	if nameLen != helloProgramNameLen {
		return false
	}

	for i := 0; i < helloProgramNameLen; i++ {
		if name[i] != helloProgramName[i] {
			return false
		}
	}

	return true
}
