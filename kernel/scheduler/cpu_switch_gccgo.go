//go:build gccgo

package scheduler

// CpuSwitch is implemented in assembly and linked by the Makefile
func CpuSwitch(oldESP *uint64, newESP uint64)

func cpuSwitch(oldESP *uint64, newESP uint64) {
	CpuSwitch(oldESP, newESP)
}
