package shell

import (
	"unsafe"

	"github.com/dmarro89/go-dav-os/terminal"
)

const (
	prompt  = "> "
	maxLine = 128
)

var (
	lineBuf  [maxLine]byte
	lineLen  int
	getTicks func() uint64
)

func SetTickProvider(fn func() uint64) { getTicks = fn }

func Init() {
	lineLen = 0
	terminal.Print(prompt)
}

func FeedRune(r rune) {
	if r == '\r' {
		r = '\n'
	}

	switch r {
	case '\b':
		if lineLen == 0 {
			return
		}
		lineLen--
		terminal.Backspace()
		return

	case '\n':
		terminal.PutRune('\n')
		execute()
		lineLen = 0
		terminal.Print(prompt)
		return
	}

	if r < 32 || r > 126 {
		return
	}
	if lineLen >= maxLine {
		return
	}

	lineBuf[lineLen] = byte(r)
	lineLen++
	terminal.PutRune(r)
}

func execute() {
	start := trimLeft(0, lineLen)
	end := trimRight(start, lineLen)
	if start >= end {
		return
	}

	cmdStart, cmdEnd := firstToken(start, end)

	if matchLiteral(cmdStart, cmdEnd, "help") {
		terminal.Print("Commands: help, clear, echo, ticks, mem\n")
		return
	}

	if matchLiteral(cmdStart, cmdEnd, "clear") {
		terminal.Clear()
		return
	}

	if matchLiteral(cmdStart, cmdEnd, "echo") {
		msgStart := trimLeft(cmdEnd, end)
		if msgStart < end {
			printRange(msgStart, end)
		}
		terminal.PutRune('\n')
		return
	}

	if matchLiteral(cmdStart, cmdEnd, "ticks") {
		if getTicks == nil {
			terminal.Print("ticks: not wired yet\n")
			return
		}
		printUint(getTicks())
		terminal.PutRune('\n')
		return
	}

	// VGA mem 0xB8000 160
	// kernel mem 0x00100000 256, mem 0x00101000 256 ...
	// .rodata & .data mem 0x00104000 256, mem 0x00108000 256, mem 0x0010C000 256
	if matchLiteral(cmdStart, cmdEnd, "mem") {
		a1s, a1e, ok := nextArg(cmdEnd, end)
		if !ok {
			terminal.Print("Usage: mem <hex_addr> [len]\n")
			return
		}

		addr, ok := parseHex32(a1s, a1e)
		if !ok {
			terminal.Print("mem: invalid hex address\n")
			return
		}

		length := 64
		a2s, a2e, ok := nextArg(a1e, end)
		if ok {
			v, ok2 := parseDec(a2s, a2e)
			if !ok2 {
				terminal.Print("mem: invalid length\n")
				return
			}
			length = v
		}

		if length < 1 {
			length = 1
		}
		if length > 512 {
			length = 512
		}

		dumpMemory(addr, length)
		return
	}

	terminal.Print("Unknown command: ")
	printRange(cmdStart, cmdEnd)
	terminal.PutRune('\n')
}

func isSpace(b byte) bool {
	return b == ' ' || b == '\t'
}

func trimLeft(start, end int) int {
	i := start
	for i < end && i < maxLine && isSpace(lineBuf[i]) {
		i++
	}
	return i
}

func trimRight(start, end int) int {
	i := end
	for i > start && i-1 < maxLine && isSpace(lineBuf[i-1]) {
		i--
	}
	return i
}

func firstToken(start, end int) (int, int) {
	i := start
	for i < end && i < maxLine && !isSpace(lineBuf[i]) {
		i++
	}
	return start, i
}

func matchLiteral(start, end int, lit string) bool {
	if end-start != len(lit) {
		return false
	}
	for i := 0; i < len(lit); i++ {
		pos := start + i
		if pos < 0 || pos >= maxLine {
			return false
		}
		if lineBuf[pos] != lit[i] {
			return false
		}
	}
	return true
}

func printRange(start, end int) {
	i := start
	for i < end && i < maxLine {
		terminal.PutRune(rune(lineBuf[i]))
		i++
	}
}

func printUint(v uint64) {
	if v == 0 {
		terminal.PutRune('0')
		return
	}

	var buf [20]byte
	i := 20
	for v > 0 {
		i--
		buf[i] = byte('0' + (v % 10))
		v /= 10
	}

	for j := i; j < 20; j++ {
		terminal.PutRune(rune(buf[j]))
	}
}

func nextArg(start, end int) (int, int, bool) {
	i := trimLeft(start, end)
	if i >= end {
		return 0, 0, false
	}
	s, e := firstToken(i, end)
	if s >= e {
		return 0, 0, false
	}
	return s, e, true
}

func parseDec(start, end int) (int, bool) {
	if start >= end {
		return 0, false
	}
	n := 0
	for i := start; i < end; i++ {
		c := lineBuf[i]
		if c < '0' || c > '9' {
			return 0, false
		}
		n = n*10 + int(c-'0')
	}
	return n, true
}

func parseHex32(start, end int) (uint32, bool) {
	if start >= end {
		return 0, false
	}
	if end-start >= 2 && lineBuf[start] == '0' && (lineBuf[start+1] == 'x' || lineBuf[start+1] == 'X') {
		start += 2
	}
	if start >= end {
		return 0, false
	}

	var v uint32
	for i := start; i < end; i++ {
		c := lineBuf[i]
		var d byte
		switch {
		case c >= '0' && c <= '9':
			d = c - '0'
		case c >= 'a' && c <= 'f':
			d = c - 'a' + 10
		case c >= 'A' && c <= 'F':
			d = c - 'A' + 10
		default:
			return 0, false
		}
		v = (v << 4) | uint32(d)
	}
	return v, true
}

func dumpMemory(addr uint32, length int) {
	off := 0
	for off < length {
		printHex32(addr + uint32(off))
		terminal.Print(": ")

		for j := 0; j < 16; j++ {
			if off+j < length {
				b := *(*byte)(unsafe.Pointer(uintptr(addr) + uintptr(off+j)))
				printHex8(b)
				terminal.PutRune(' ')
			} else {
				terminal.Print("   ")
			}
		}

		terminal.Print(" |")

		for j := 0; j < 16; j++ {
			if off+j < length {
				b := *(*byte)(unsafe.Pointer(uintptr(addr) + uintptr(off+j)))
				if b >= 32 && b <= 126 {
					terminal.PutRune(rune(b))
				} else {
					terminal.PutRune('.')
				}
			} else {
				terminal.PutRune(' ')
			}
		}

		terminal.Print("|\n")
		off += 16
	}
}

func printHex32(v uint32) {
	hexDigits := "0123456789ABCDEF"
	for i := 7; i >= 0; i-- {
		n := byte((v >> (uint(i) * 4)) & 0xF)
		terminal.PutRune(rune(hexDigits[n]))
	}
}

func printHex8(b byte) {
	hexDigits := "0123456789ABCDEF"
	terminal.PutRune(rune(hexDigits[(b>>4)&0xF]))
	terminal.PutRune(rune(hexDigits[b&0xF]))
}
