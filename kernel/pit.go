package kernel

const pitFreq = 1193182

func PITInit(hz uint32) {
	if hz == 0 {
		hz = 100
	}
	div := uint16(pitFreq / hz)

	// channel 0, lobyte/hibyte, mode 3 (square wave), binary
	outb(0x43, 0x36)
	outb(0x40, byte(div&0xFF))
	outb(0x40, byte(div>>8))
}
