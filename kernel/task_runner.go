package kernel

import "github.com/dmarro89/go-dav-os/kernel/scheduler"

var helloProgramMsg = [...]byte{
	'H', 'e', 'l', 'l', 'o', ' ', 'f', 'r', 'o', 'm', ' ',
	't', 'a', 's', 'k', '\n',
}

var helloProgramName = [...]byte{'h', 'e', 'l', 'l', 'o'}

const helloProgramNameLen = 5

func RunProgram(name *[16]byte, nameLen int) (pid int, ok bool) {
	if nameLen != helloProgramNameLen {
		return -1, false
	}

	for i := 0; i < helloProgramNameLen; i++ {
		if name[i] != helloProgramName[i] {
			return -1, false
		}
	}

	task := scheduler.NewTask(programHello)
	if task == nil {
		return -1, false
	}
	return task.ID, true
}

func programHello() {
	TriggerSysWrite(&helloProgramMsg[0], uint32(len(helloProgramMsg)))
	TriggerSysExit(0)
	for {
	}
}
