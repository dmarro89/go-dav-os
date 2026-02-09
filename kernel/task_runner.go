package kernel

import "github.com/dmarro89/go-dav-os/kernel/scheduler"

type builtInProgram struct {
	name    [16]byte
	nameLen int
	entry   func()
}

var helloProgramMsg = [...]byte{
	'H', 'e', 'l', 'l', 'o', ' ', 'f', 'r', 'o', 'm', ' ',
	't', 'a', 's', 'k', '\n',
}

var builtInPrograms = [...]builtInProgram{
	{
		name:    [16]byte{'h', 'e', 'l', 'l', 'o'},
		nameLen: 5,
		entry:   programHello,
	},
}

func RunProgram(name *[16]byte, nameLen int) (pid int, ok bool) {
	for i := 0; i < len(builtInPrograms); i++ {
		p := &builtInPrograms[i]
		if p.nameLen != nameLen {
			continue
		}

		match := true
		for j := 0; j < nameLen; j++ {
			if p.name[j] != name[j] {
				match = false
				break
			}
		}
		if !match {
			continue
		}

		task := scheduler.NewTask(p.entry)
		if task == nil {
			return -1, false
		}
		return task.ID, true
	}

	return -1, false
}

func programHello() {
	TriggerSysWrite(&helloProgramMsg[0], uint32(len(helloProgramMsg)))
	TriggerSysExit(0)
	for {
	}
}
