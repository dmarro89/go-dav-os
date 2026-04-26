//go:build !testing

package kernel

const (
	kernelCodeSelector uint16 = 0x08
	kernelDataSelector uint16 = 0x10
	userCodeSelector   uint16 = 0x1B
	userDataSelector   uint16 = 0x23
	tssSelector        uint16 = 0x28
)
