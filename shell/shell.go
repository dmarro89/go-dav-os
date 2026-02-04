package shell

import (
	"unsafe"

	"github.com/dmarro89/go-dav-os/drivers/ata"
	"github.com/dmarro89/go-dav-os/fs"
	"github.com/dmarro89/go-dav-os/fs/fat16"
	"github.com/dmarro89/go-dav-os/mem"
	"github.com/dmarro89/go-dav-os/terminal"
)

const (
	prompt    = "> "
	maxLine   = 128
	osName    = "DavOS"
	maxDistanceThreshold = 3
	osVersion = "0.2.0"
)

var (
	lineBuf  [maxLine]byte
	lineLen  int
	getTicks func() uint64
	tmpName  [16]byte
	tmpData  [4096]byte
	diskBuf  [512]byte

	// History ring buffer
	// historyBuf stores the content of the commands
	historyBuf [32][maxLine]byte
	// historyLen stores the length of each command in the buffer
	historyLen [32]int
	// historyHead points to the next free slot in the ring buffer
	historyHead int
	// historyCount tracks the total number of items currently stored (max 32)
	historyCount int
)

// maxHistory defines the maximum size of the history ring buffer
const maxHistory = 32

var commandBuf = [...]string{
	"help", "clear", "echo", "ticks", "mem", "mmap",
	"pfa", "alloc", "free", "ls", "write", "cat", "rm", "stat",
	"version", "history",
}

func SetTickProvider(fn func() uint64) { getTicks = fn }

func Init() {
	lineLen = 0
	terminal.Print("Welcome to " + osName + " " + osVersion + "\n")
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

	// Add to history
	// We only add non-empty commands to history.
	// We also avoid adding duplicate consecutive commands.
	// The buffer is implemented as a standard ring buffer, overwriting old entries
	// when the buffer is full.
	histLen := end - start
	if histLen > 0 {
		// duplicate check
		isDuplicate := false
		if historyCount > 0 {
			lastIdx := (historyHead - 1 + maxHistory) % maxHistory
			if historyLen[lastIdx] == histLen {
				match := true
				for i := 0; i < histLen; i++ {
					if historyBuf[lastIdx][i] != lineBuf[start+i] {
						match = false
						break
					}
				}
				if match {
					isDuplicate = true
				}
			}
		}

		if !isDuplicate {
			idx := historyHead
			for i := 0; i < histLen; i++ {
				historyBuf[idx][i] = lineBuf[start+i]
			}
			historyLen[idx] = histLen

			historyHead = (historyHead + 1) % maxHistory
			if historyCount < maxHistory {
				historyCount++
			}
		}
	}

	cmdStart, cmdEnd := firstToken(start, end)

	if matchLiteral(cmdStart, cmdEnd, "history") {
		// Print history
		// We iterate from 0 to historyCount-1.
		// To print chronologically (oldest first), we calculate the starting index
		// in the ring buffer: (head - count + size) % size.
		startIdx := (historyHead - historyCount + maxHistory) % maxHistory
		for i := 0; i < historyCount; i++ {
			idx := (startIdx + i) % maxHistory

			// Print index (1-based for user friendliness)
			printUint(uint64(i + 1))
			terminal.Print(" ")

			// Print command content
			l := historyLen[idx]
			for k := 0; k < l; k++ {
				terminal.PutRune(rune(historyBuf[idx][k]))
			}
			terminal.PutRune('\n')
		}
		return
	}

	if matchLiteral(cmdStart, cmdEnd, "help") {
		terminal.Print("Commands: help, clear, echo, ticks, mem, mmap, pfa, alloc, free, ls, write, cat, rm, stat, version, history, disk, fatinit, fatformat, fatinfo, fatls, fatcreate, fatread\n")
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

		addr, ok := parseHex64(a1s, a1e)
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

	if matchLiteral(cmdStart, cmdEnd, "mmapmax") {
		var maxEnd uint64
		n := mem.MMapCount()
		for i := 0; i < n; i++ {
			bLo, bHi, lLo, lHi, typ := mem.MMapEntry(i)
			if typ != 1 {
				continue
			}
			base := (uint64(bHi) << 32) | uint64(bLo)
			l := (uint64(lHi) << 32) | uint64(lLo)
			end := base + l
			if end > maxEnd {
				maxEnd = end
			}
		}

		terminal.Print("mmap max end=0x")
		printHexU64(maxEnd)
		terminal.PutRune('\n')
		return
	}

	if matchLiteral(cmdStart, cmdEnd, "pfa") {
		if !mem.PFAReady() {
			terminal.Print("pfa: not ready\n")
			return
		}

		terminal.Print("pages total=")
		printUint(mem.TotalPages())
		terminal.Print(" used=")
		printUint(mem.UsedPages())
		terminal.Print(" free=")
		printUint(mem.FreePages())
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
		printHexU64(addr)
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

		addr, ok := parseHex64(a1s, a1e)
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
			printUint(size)
			terminal.Print("  page=0x")
			printHexU64(page)
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
		for i := uint64(0); i < size; i++ {
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
		printHexU64(page)
		terminal.Print(" size=")
		printUint(size)
		terminal.PutRune('\n')
		return
	}

	if matchLiteral(cmdStart, cmdEnd, "disk") {
		a1s, a1e, ok := nextArg(cmdEnd, end)
		if !ok {
			terminal.Print("Usage: disk <read|write> <lba> [text]\n")
			return
		}

		if matchLiteral(a1s, a1e, "read") {
			a2s, a2e, ok := nextArg(a1e, end)
			if !ok {
				terminal.Print("disk read: missing lba\n")
				return
			}

			lba := 0
			// Try hex then dec
			vHex, okHex := parseHex64(a2s, a2e)
			if okHex {
				lba = int(vHex)
			} else {
				vDec, okDec := parseDec(a2s, a2e)
				if !okDec {
					terminal.Print("disk read: invalid lba\n")
					return
				}
				lba = vDec
			}

			if ata.ReadSector(uint32(lba), &diskBuf) {
				terminal.Print("Read Sector ")
				printUint(uint64(lba))
				terminal.Print(" OK\n")
				dumpMemory(uint64(uintptr(unsafe.Pointer(&diskBuf[0]))), 512)
			} else {
				terminal.Print("Read Failed\n")
			}
			return
		}

		if matchLiteral(a1s, a1e, "write") {
			a2s, a2e, ok := nextArg(a1e, end)
			if !ok {
				terminal.Print("disk write: missing lba\n")
				return
			}
			lba, ok := parseDec(a2s, a2e)
			if !ok {
				// try hex if needed, but dec is fine
				vHex, okHex := parseHex64(a2s, a2e)
				if okHex {
					lba = int(vHex)
				} else {
					terminal.Print("disk write: invalid lba\n")
					return
				}
			}

			msgStart := trimLeft(a2e, end)
			// Clear diskBuf before writing to avoid stale data
			for i := 0; i < 512; i++ {
				diskBuf[i] = 0
			}
			idx := 0
			for i := msgStart; i < end && idx < 512; i++ {
				diskBuf[idx] = lineBuf[i]
				idx++
			}

			if ata.WriteSector(uint32(lba), &diskBuf) {
				terminal.Print("Write Sector ")
				printUint(uint64(lba))
				terminal.Print(" OK\n")
			} else {
				terminal.Print("Write Failed\n")
			}
			return
		}
	}

	if matchLiteral(cmdStart, cmdEnd, "fatinit") {
		if fat16.Init() {
			terminal.Print("FAT16 Initialized\n")
		} else {
			terminal.Print("FAT16 Init Failed\n")
		}
		return
	}

	if matchLiteral(cmdStart, cmdEnd, "fatformat") {
		if fat16.Format() {
			terminal.Print("FAT16 Formatted\n")
		} else {
			terminal.Print("FAT16 Format Failed\n")
		}
		return
	}

	if matchLiteral(cmdStart, cmdEnd, "fatinfo") {
		fat16.Info()
		return
	}

	if matchLiteral(cmdStart, cmdEnd, "fatls") {
		fat16.ListDir()
		return
	}

	if matchLiteral(cmdStart, cmdEnd, "fatcreate") {
		// Usage: fatcreate <filename> <content>
		a1s, a1e, ok := nextArg(cmdEnd, end)
		if !ok {
			terminal.Print("Usage: fatcreate <filename> <content>\n")
			return
		}

		// Parse filename (max 8 chars, no extension for simplicity)
		var fname [8]byte
		var fext [3]byte
		for i := 0; i < 8; i++ {
			fname[i] = ' '
		}
		for i := 0; i < 3; i++ {
			fext[i] = ' '
		}

		nameLen := a1e - a1s
		if nameLen > 8 {
			nameLen = 8
		}
		for i := 0; i < nameLen; i++ {
			c := lineBuf[a1s+i]
			if c >= 'a' && c <= 'z' {
				c = c - 'a' + 'A' // Uppercase
			}
			fname[i] = c
		}

		// Get content
		msgStart := trimLeft(a1e, end)
		var dataBuf [512]byte
		// Clear dataBuf before writing to avoid stale data
		for i := 0; i < 512; i++ {
			dataBuf[i] = 0
		}
		idx := 0
		for i := msgStart; i < end && idx < 512; i++ {
			dataBuf[idx] = lineBuf[i]
			idx++
		}

		if fat16.CreateFile(&fname, &fext, &dataBuf, uint32(idx)) {
			terminal.Print("File created\n")
		} else {
			terminal.Print("Failed to create file\n")
		}
		return
	}

	if matchLiteral(cmdStart, cmdEnd, "fatread") {
		// Usage: fatread <filename>
		a1s, a1e, ok := nextArg(cmdEnd, end)
		if !ok {
			terminal.Print("Usage: fatread <filename>\n")
			return
		}

		var fname [8]byte
		var fext [3]byte
		for i := 0; i < 8; i++ {
			fname[i] = ' '
		}
		for i := 0; i < 3; i++ {
			fext[i] = ' '
		}

		nameLen := a1e - a1s
		if nameLen > 8 {
			nameLen = 8
		}
		for i := 0; i < nameLen; i++ {
			c := lineBuf[a1s+i]
			if c >= 'a' && c <= 'z' {
				c = c - 'a' + 'A'
			}
			fname[i] = c
		}

		size, ok := fat16.ReadFile(&fname, &fext, &diskBuf)
		if !ok {
			terminal.Print("File not found\n")
			return
		}

		for i := uint32(0); i < size && i < 512; i++ {
			terminal.PutRune(rune(diskBuf[i]))
		}
		terminal.PutRune('\n')
		return
	}

	if matchLiteral(cmdStart, cmdEnd, "version") {
		terminal.Print(osName + " " + osVersion)
		proof := uint64(0x0123456789ABCDEF)
		if (proof >> 32) != 0 {
			terminal.Print(" (64bit)")
		}
		terminal.PutRune('\n')
		return
	}

	var suggestionBuf [len(commandBuf)]string
	suggestionCount := 0

	for i := 0; i < len(commandBuf); i++ {
		if calculateDistance(cmdStart, cmdEnd, commandBuf[i]) < maxDistanceThreshold {
			suggestionBuf[suggestionCount] = commandBuf[i]
			suggestionCount++
		}
	}

	if suggestionCount > 0 {
		terminal.Print("Did you mean '")
		for i := 0; i < suggestionCount; i++ {
			if i > 0 { // To ensure formatting; the first time we print we do not add a space in-front of the suggestion
				terminal.Print(" ")
			}
			terminal.Print(suggestionBuf[i])
		}
		terminal.Print("'?")
		terminal.PutRune('\n')
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
	if (v >> 32) != 0 {
		terminal.Print("0x")
		printHexU64(v)
		return
	}

	u := uint32(v)
	var buf [10]byte
	i := 10
	for u > 0 {
		i--
		buf[i] = byte('0' + (u % 10))
		u /= 10
	}

	for j := i; j < 10; j++ {
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

func parseHex64(start, end int) (uint64, bool) {
	if start >= end {
		return 0, false
	}
	if end-start >= 2 && lineBuf[start] == '0' && (lineBuf[start+1] == 'x' || lineBuf[start+1] == 'X') {
		start += 2
	}
	if start >= end {
		return 0, false
	}

	var v uint64
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
		v = (v << 4) | uint64(d)
	}
	return v, true
}

func dumpMemory(addr uint64, length int) {
	off := 0
	for off < length {
		printHexU64(addr + uint64(off))
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

func printHexU64(v uint64) {
	const hexDigits = "0123456789ABCDEF"
	for i := 15; i >= 0; i-- {
		n := byte((v >> (uint(i) * 4)) & 0xF)
		terminal.PutRune(rune(hexDigits[n]))
	}
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

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func calculateDistance(start, end int, cmd string) int {
	lenA := end - start
	lenB := len(cmd)

	var prev [maxLine + 1]int
	var curr [maxLine + 1]int

	for j := 0; j <= lenB; j++ {
		prev[j] = j
	}

	for i := 1; i <= lenA; i++ {
		curr[0] = i
		for j := 1; j <= lenB; j++ {
			cost := 1
			if lineBuf[start+i-1] == cmd[j-1] {
				cost = 0
			}
			curr[j] = minInt(
				minInt(prev[j]+1, curr[j-1]+1),
				prev[j-1]+cost,
			)
		}
		for j := 0; j <= lenB; j++ {
			prev[j] = curr[j]
		}
	}

	return prev[lenB]
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
