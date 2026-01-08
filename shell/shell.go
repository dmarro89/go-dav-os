package shell

import (
	"unsafe"

	"github.com/dmarro89/go-dav-os/fs"
	"github.com/dmarro89/go-dav-os/mem"
	"github.com/dmarro89/go-dav-os/terminal"
)

const (
	prompt    = "> "
	maxLine   = 128
	osName    = "DavOS"
	osVersion = "0.1.0"
)

var (
	lineBuf  [maxLine]byte
	lineLen  int
	getTicks func() uint64
	tmpName  [16]byte
	tmpData  [4096]byte
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
		terminal.Print("Commands: help, clear, echo, ticks, mem, mmap, pfa, alloc, free, ls, write, cat, rm, stat, version\n")
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

	if matchLiteral(cmdStart, cmdEnd, "mmap") {
		n := mem.MMapCount()
		for i := 0; i < n; i++ {
			bLo, bHi, lLo, lHi, typ := mem.MMapEntry(i)

			terminal.Print("base=0x")
			printHex64(bHi, bLo)
			terminal.Print(" len=0x")
			printHex64(lHi, lLo)
			terminal.Print(" type=")
			printUint(uint64(typ))
			terminal.PutRune('\n')
		}
		return
	}

	if matchLiteral(cmdStart, cmdEnd, "pfa") {
		if !mem.PFAReady() {
			terminal.Print("pfa: not ready\n")
			return
		}

		terminal.Print("pages total=")
		printUint(uint64(mem.TotalPages()))
		terminal.Print(" used=")
		printUint(uint64(mem.UsedPages()))
		terminal.Print(" free=")
		printUint(uint64(mem.FreePages()))
		terminal.PutRune('\n')
		return
	}

	if matchLiteral(cmdStart, cmdEnd, "alloc") {
		// allocate one 4KB page and print its physical address
		if !mem.PFAReady() {
			terminal.Print("alloc: pfa not ready\n")
			return
		}

		addr := mem.AllocPage()
		if addr == 0 {
			terminal.Print("alloc: failed\n")
			return
		}

		terminal.Print("0x")
		printHex32(addr)
		terminal.PutRune('\n')
		return
	}

	if matchLiteral(cmdStart, cmdEnd, "free") {
		// free a previously allocated 4KB page
		if !mem.PFAReady() {
			terminal.Print("free: pfa not ready\n")
			return
		}

		a1s, a1e, ok := nextArg(cmdEnd, end)
		if !ok {
			terminal.Print("Usage: free <hex_addr>\n")
			return
		}

		addr, ok := parseHex32(a1s, a1e)
		if !ok {
			terminal.Print("free: invalid hex address\n")
			return
		}

		if mem.FreePage(addr) {
			terminal.Print("ok\n")
		} else {
			terminal.Print("free: failed\n")
		}
		return
	}

	if matchLiteral(cmdStart, cmdEnd, "ls") {
		for i := 0; i < fs.MaxFiles(); i++ {
			used, name, nameLen, size, page := fs.Entry(i)
			if !used {
				continue
			}

			printName(name, nameLen)
			terminal.Print("  size=")
			printUint(uint64(size))
			terminal.Print("  page=0x")
			printHex32(page)
			terminal.PutRune('\n')
		}
		return
	}

	if matchLiteral(cmdStart, cmdEnd, "write") {
		a1s, a1e, ok := nextArg(cmdEnd, end)
		if !ok {
			terminal.Print("Usage: write <name> <text...>\n")
			return
		}

		nameLen, ok := copyNameFromRange(a1s, a1e)
		if !ok {
			terminal.Print("write: invalid name\n")
			return
		}

		msgStart := trimLeft(a1e, end)
		dataLen := copyDataFromRange(msgStart, end)

		if !fs.Write(&tmpName, nameLen, (*byte)(unsafe.Pointer(&tmpData[0])), dataLen) {
			terminal.Print("write: failed\n")
			return
		}

		terminal.Print("ok\n")
		return
	}

	if matchLiteral(cmdStart, cmdEnd, "cat") {
		a1s, a1e, ok := nextArg(cmdEnd, end)
		if !ok {
			terminal.Print("Usage: cat <name>\n")
			return
		}

		nameLen, ok := copyNameFromRange(a1s, a1e)
		if !ok {
			terminal.Print("cat: invalid name\n")
			return
		}

		page, size, ok := fs.Lookup(&tmpName, nameLen)
		if !ok {
			terminal.Print("cat: not found\n")
			return
		}

		p := uintptr(page)
		for i := uint32(0); i < size; i++ {
			b := *(*byte)(unsafe.Pointer(p + uintptr(i)))
			terminal.PutRune(rune(b))
		}
		terminal.PutRune('\n')
		return
	}

	if matchLiteral(cmdStart, cmdEnd, "rm") {
		a1s, a1e, ok := nextArg(cmdEnd, end)
		if !ok {
			terminal.Print("Usage: rm <name>\n")
			return
		}

		nameLen, ok := copyNameFromRange(a1s, a1e)
		if !ok {
			terminal.Print("rm: invalid name\n")
			return
		}

		if fs.Remove(&tmpName, nameLen) {
			terminal.Print("ok\n")
		} else {
			terminal.Print("rm: not found\n")
		}
		return
	}

	if matchLiteral(cmdStart, cmdEnd, "stat") {
		a1s, a1e, ok := nextArg(cmdEnd, end)
		if !ok {
			terminal.Print("Usage: stat <name>\n")
			return
		}

		nameLen, ok := copyNameFromRange(a1s, a1e)
		if !ok {
			terminal.Print("stat: invalid name\n")
			return
		}

		page, size, ok := fs.Lookup(&tmpName, nameLen)
		if !ok {
			terminal.Print("stat: not found\n")
			return
		}

		terminal.Print("page=0x")
		printHex32(page)
		terminal.Print(" size=")
		printUint(uint64(size))
		terminal.PutRune('\n')
		return
	}

	if matchLiteral(cmdStart, cmdEnd, "version") {
		terminal.Print(osName + " " + osVersion + "\n")
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

func printHex64(hi, lo uint32) {
	printHex32(hi)
	printHex32(lo)
}

func printHex8(b byte) {
	hexDigits := "0123456789ABCDEF"
	terminal.PutRune(rune(hexDigits[(b>>4)&0xF]))
	terminal.PutRune(rune(hexDigits[b&0xF]))
}

func printName(name *[16]byte, nameLen int) {
	for i := 0; i < nameLen; i++ {
		terminal.PutRune(rune(name[i]))
	}
}

func copyNameFromRange(start, end int) (int, bool) {
	n := end - start
	if n <= 0 || n > 16 {
		return 0, false
	}
	for i := 0; i < 16; i++ {
		tmpName[i] = 0
	}
	for i := 0; i < n; i++ {
		tmpName[i] = lineBuf[start+i]
	}
	return n, true
}

func copyDataFromRange(start, end int) uint32 {
	if end < start {
		return 0
	}
	n := end - start
	if n < 0 {
		return 0
	}
	if n > 4096 {
		n = 4096
	}
	for i := 0; i < n; i++ {
		tmpData[i] = lineBuf[start+i]
	}
	return uint32(n)
}
