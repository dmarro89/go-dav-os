package kernel

import (
	"github.com/dmarro89/go-dav-os/kernel/scheduler"
	"github.com/dmarro89/go-dav-os/keyboard"
)

var ticks uint64

func IRQ0Handler() {
	ticks++
	PICEOI(0)
	scheduler.Schedule()
}

func IRQ1Handler() {
	// Read & buffer scancode -> rune (no terminal printing here!)
	keyboard.IRQHandler()

	// Tell PIC we're done with IRQ1, otherwise it won't fire again
	PICEOI(1)
}

func GetTicks() uint64 {
	return ticks
}
