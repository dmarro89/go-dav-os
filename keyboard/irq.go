package keyboard

const bufSize = 256

var buf [bufSize]rune
var head uint32
var tail uint32

func push(r rune) {
	next := (head + 1) & (bufSize - 1)
	if next == tail {
		// buffer full -> drop char
		return
	}
	buf[head] = r
	head = next
}

// Called from kernel IRQ1 handler
func IRQHandler() {
	sc := inb(0x60)

	// Ignore key releases (highest bit set)
	if sc&0x80 != 0 {
		return
	}

	if int(sc) >= len(LayoutIT) {
		return
	}

	r := LayoutIT[sc]
	if r == 0 {
		return
	}

	push(r)
}

// Non-blocking read used by the shell loop.
func TryRead() (rune, bool) {
	if tail == head {
		return 0, false
	}
	r := buf[tail]
	tail = (tail + 1) & (bufSize - 1)
	return r, true
}
