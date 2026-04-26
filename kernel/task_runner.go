//go:build !testing

package kernel

var helloProgramName = [...]byte{'h', 'e', 'l', 'l', 'o'}
var kernelReadProbeProgramName = [...]byte{'k', 'r', 'e', 'a', 'd'}
var kernelWriteProbeProgramName = [...]byte{'k', 'w', 'r', 'i', 't', 'e'}

func ExecuteUserTask(rip, rsp uint64)
func GetUserProgramHelloAddr() uint64
func GetUserProgramKernelReadProbeAddr() uint64
func GetUserProgramKernelWriteProbeAddr() uint64
func GetUserStackTopAddr() uint64

func RunProgram(name *[16]byte, nameLen int) (pid int, ok bool) {
	var rip uint64
	switch {
	case matchProgramName(name, nameLen, helloProgramName[:]):
		rip = GetUserProgramHelloAddr()
	case matchProgramName(name, nameLen, kernelReadProbeProgramName[:]):
		rip = GetUserProgramKernelReadProbeAddr()
	case matchProgramName(name, nameLen, kernelWriteProbeProgramName[:]):
		rip = GetUserProgramKernelWriteProbeAddr()
	default:
		return -1, false
	}

	ExecuteUserTask(rip, GetUserStackTopAddr())

	return 1, true
}

func matchProgramName(name *[16]byte, nameLen int, expected []byte) bool {
	if nameLen != len(expected) {
		return false
	}

	for i := 0; i < len(expected); i++ {
		if name[i] != expected[i] {
			return false
		}
	}
	return true
}
