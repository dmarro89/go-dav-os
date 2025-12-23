package kernel

import (
	"github.com/dmarro89/go-dav-os/keyboard"
)

var ticks uint32

func IRQ0Handler() {
	ticks++
	PICEOI(0)
}

func IRQ1Handler() {
	// Read & buffer scancode -> rune (no terminal printing here!)
	keyboard.IRQHandler()

	// Tell PIC we're done with IRQ1, otherwise it won't fire again
	PICEOI(1)
}
