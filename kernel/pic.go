package kernel

const (
	pic1Cmd  = 0x20
	pic1Data = 0x21
	pic2Cmd  = 0xA0
	pic2Data = 0xA1

	icw1Init  = 0x11
	icw4_8086 = 0x01

	eoi = 0x20
)

func PICRemap(offset1, offset2 byte) {
	// start init
	outb(pic1Cmd, icw1Init)
	outb(pic2Cmd, icw1Init)

	// set vector offsets
	outb(pic1Data, offset1) // master offset (0x20)
	outb(pic2Data, offset2) // slave  offset (0x28)

	// tell Master about Slave at IRQ2, tell Slave its cascade identity
	outb(pic1Data, 0x04)
	outb(pic2Data, 0x02)

	// 8086 mode
	outb(pic1Data, icw4_8086)
	outb(pic2Data, icw4_8086)
}

func PICSetMask(masterMask, slaveMask byte) {
	outb(pic1Data, masterMask)
	outb(pic2Data, slaveMask)
}

func PICEOI(irq byte) {
	// If the IRQ came from the slave PIC, we need to notify it first
	if irq >= 8 {
		outb(pic2Cmd, eoi)
	}
	outb(pic1Cmd, eoi)
}
