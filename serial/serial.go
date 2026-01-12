package serial

const com1Base uint16 = 0x3F8

func inb(port uint16) byte
func outb(port uint16, value byte)

var enabled bool

func Init() {
	outb(com1Base+1, 0x00) // disable interrupts
	outb(com1Base+3, 0x80) // enable DLAB
	outb(com1Base+0, 0x01) // divisor low (115200 baud)
	outb(com1Base+1, 0x00) // divisor high
	outb(com1Base+3, 0x03) // 8N1
	outb(com1Base+2, 0xC7) // enable FIFO, clear, 14-byte threshold
	outb(com1Base+4, 0x0B) // IRQs enabled, RTS/DSR set
	enabled = true
}

func Enabled() bool {
	return enabled
}

func WriteByte(b byte) {
	if !enabled {
		return
	}
	for (inb(com1Base+5) & 0x20) == 0 {
	}
	outb(com1Base, b)
}

func TryRead() (rune, bool) {
	if !enabled {
		return 0, false
	}
	if (inb(com1Base+5) & 0x01) == 0 {
		return 0, false
	}
	b := inb(com1Base)
	if b == '\r' {
		b = '\n'
	}
	if b == 0x7f {
		b = '\b'
	}
	return rune(b), true
}
