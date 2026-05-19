//go:build testing

package keyboard

type Layout interface {
	GetKey(byte) (rune, bool)
	GetShiftDigitSymbol(r rune) (rune, bool)
}

func inb(port uint16) byte {
	return 0
}

func outb(port uint16, value byte) {}

func IRQHandler() {}

func TryRead() (rune, bool) {
	return 0, false
}
