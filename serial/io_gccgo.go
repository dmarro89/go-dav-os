//go:build gccgo

package serial

const (
	com1Base       uint16 = 0x3F8
	lineStatusPort        = com1Base + 5
	dataReady      byte   = 1
	transmitReady  byte   = 1 << 5
)

func inb(port uint16) byte
func outb(port uint16, value byte)

func initPort() bool {
	outb(com1Base+1, 0x00)
	outb(com1Base+3, 0x80)
	outb(com1Base, 0x03)
	outb(com1Base+1, 0x00)
	outb(com1Base+3, 0x03)
	outb(com1Base+2, 0xC7)
	outb(com1Base+4, 0x1E)
	outb(com1Base, 0xAE)
	if inb(com1Base) != 0xAE {
		return false
	}
	outb(com1Base+4, 0x0F)
	return true
}

func tryWriteByte(value byte) bool {
	if inb(lineStatusPort)&transmitReady == 0 {
		return false
	}
	outb(com1Base, value)
	return true
}

func tryReadByte() (byte, bool) {
	if inb(lineStatusPort)&dataReady == 0 {
		return 0, false
	}
	return inb(com1Base), true
}
