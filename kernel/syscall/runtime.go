//go:build !testing

package syscall

func Init(readMSR func(uint32) uint64, writeMSR func(uint32, uint64), lstar uint64, kernelCS, userCS uint16) {
	writeMSR(MSRSTAR, STARValue(kernelCS, userCS))
	writeMSR(MSRLSTAR, lstar)
	writeMSR(MSRSFMASK, SFMASKValue())
	writeMSR(MSREFER, EnableSCE(readMSR(MSREFER)))
}
