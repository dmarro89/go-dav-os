//go:build !testing

package keyboard

func inb(port uint16) byte
func outb(port uint16, value byte)

const (
	portData   uint16 = 0x60
	portStatus uint16 = 0x64

	statusOutputBuffer = 1 // bit 0 => output buffer full
)

const (
	scLeftShiftDown  = 0x2A
	scLeftShiftUp    = 0xAA
	scRightShiftDown = 0x36
	scRightShiftUp   = 0xB6
	scCapsLockDown   = 0x3A
)

var (
	leftShiftDown  bool
	rightShiftDown bool
	capsLockOn     bool
)

var currentLayout Layout

func SetLayout(l Layout) {
	currentLayout = l
}

func readScancode() byte {
	for {
		status := inb(portStatus)
		if (status & statusOutputBuffer) != 0 {
			return inb(portData)
		}
	}
}

func translateScancode(sc byte) (rune, bool) {
	if handleModifier(sc) {
		return 0, false
	}

	// Ignore key releases for non-modifier keys
	if sc&0x80 != 0 {
		return 0, false
	}

	r, valid := currentLayout.GetKey(sc)
	if !valid {
		return 0, false
	}

	if isASCIILetter(r) {
		if shiftActive() != capsLockOn { // XOR between shift and caps
			r = toUpperASCII(r)
		} else {
			r = toLowerASCII(r)
		}
		return r, true
	}

	if isASCIIDigit(r) && shiftActive() {
		if sym, ok := currentLayout.GetShiftDigitSymbol(r); ok {
			return sym, true
		}
	}

	return r, true
}

func handleModifier(sc byte) bool {
	switch sc {
	case scLeftShiftDown:
		leftShiftDown = true
		return true
	case scLeftShiftUp:
		leftShiftDown = false
		return true
	case scRightShiftDown:
		rightShiftDown = true
		return true
	case scRightShiftUp:
		rightShiftDown = false
		return true
	case scCapsLockDown:
		capsLockOn = !capsLockOn
		return true
	default:
		return false
	}
}

func shiftActive() bool {
	return leftShiftDown || rightShiftDown
}

func ReadKey() rune {
	for {
		sc := readScancode()
		if r, ok := translateScancode(sc); ok {
			return r
		}
	}
}

func isASCIILetter(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

func isASCIIDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func toUpperASCII(r rune) rune {
	if r >= 'a' && r <= 'z' {
		return r - 'a' + 'A'
	}
	return r
}

func toLowerASCII(r rune) rune {
	if r >= 'A' && r <= 'Z' {
		return r - 'A' + 'a'
	}
	return r
}
