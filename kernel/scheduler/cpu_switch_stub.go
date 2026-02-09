//go:build !gccgo

package scheduler

// Stub implementation of CpuSwitch for non-gccgo builds, which simply updates the oldESP with newESP without performing an actual context switch
// This allows the scheduler to function during the tests execution
func cpuSwitch(oldESP *uint64, newESP uint64) {
	if oldESP != nil {
		*oldESP = newESP
	}
}
