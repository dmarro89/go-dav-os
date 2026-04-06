package syscall

const (
	MSREFER   uint32 = 0xC0000080
	MSRSTAR   uint32 = 0xC0000081
	MSRLSTAR  uint32 = 0xC0000082
	MSRSFMASK uint32 = 0xC0000084

	eferSCEBit   uint64 = 1 << 0
	syscallIFBit uint64 = 1 << 9
)

func STARValue(kernelCS, userCS uint16) uint64 {
	return uint64(kernelCS&^3)<<32 | uint64(userCS)<<48
}

func SFMASKValue() uint64 {
	return syscallIFBit
}

func EnableSCE(efer uint64) uint64 {
	return efer | eferSCEBit
}
