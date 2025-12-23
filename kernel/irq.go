package kernel

import "github.com/dmarro89/go-dav-os/terminal"

var ticks uint32

func IRQ0Handler() {
	ticks++

	if ticks%100 == 0 {
		terminal.Print(".")
	}

	PICEOI(0)
}

func IRQ1Handler() {
	_ = inb(0x60)

	PICEOI(1)
}
