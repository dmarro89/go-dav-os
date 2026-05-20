package shell

import (
	"unsafe"

	"github.com/dmarro89/go-dav-os/agent"
	"github.com/dmarro89/go-dav-os/drivers/ata"
	"github.com/dmarro89/go-dav-os/fs"
	"github.com/dmarro89/go-dav-os/fs/fat16"
	"github.com/dmarro89/go-dav-os/mem"
	"github.com/dmarro89/go-dav-os/terminal"
)

const (
	prompt               = "> "
	maxLine              = 128
	osName               = "DavOS"
	maxDistanceThreshold = 3
	osVersion            = "0.2.0"
	commandHelp          = "help"
	commandHistory       = "history"
)

var (
	lineBuf         [maxLine]byte
	lineLen         int
	getTicks        func() uint64
	getSyscallTicks func() uint64
	runProgram      func(name *[16]byte, nameLen int) (pid int, ok bool)
	switchLayoutFn  func(string) bool
	currentLayout   = "it"
	tmpName         [16]byte
	tmpData         [4096]byte
	diskBuf         [512]byte

	// History ring buffer
	// historyBuf stores the content of the commands
	historyBuf [32][maxLine]byte
	// historyLen stores the length of each command in the buffer
	historyLen [32]int
	// historyHead points to the next free slot in the ring buffer
	historyHead int
	// historyCount tracks the total number of items currently stored (max 32)
	historyCount int

	runtimeAgent agent.Runtime
)

// maxHistory defines the maximum size of the history ring buffer
const maxHistory = 32

var commandBuf = [...]string{
	commandHelp, commandHistory, "clear", "echo", "ticks", "uptime",
	"mem", "mmap", "mmapmax", "pfa", "alloc", "free",
	"ls", "write", "cat", "rm", "stat",
	"disk", "fatinit", "fatformat", "fatinfo", "fatls", "fatcreate", "fatread",
	"layout", "version", "run", "agent",
}

func SetTickProvider(fn func() uint64)        { getTicks = fn }
func SetSyscallTickProvider(fn func() uint64) { getSyscallTicks = fn }
func SetProgramRunner(fn func(name *[16]byte, nameLen int) (pid int, ok bool)) {
	runProgram = fn
}
func SetLayoutSwitcher(fn func(string) bool) { switchLayoutFn = fn }
func SetInitialLayout(name string)           { currentLayout = name }
func SetAgentRuntime(runtime *agent.Runtime) {
	if runtime == nil {
		runtimeAgent.Executor.ListFiles = nil
		runtimeAgent.Executor.ReadFile = nil
		runtimeAgent.Executor.WriteFile = nil
		runtimeAgent.Executor.DeleteFile = nil
		runtimeAgent.Executor.StatFile = nil
		runtimeAgent.Executor.ShowHelp = nil
		runtimeAgent.Executor.ShowHistory = nil
		runtimeAgent.Executor.ShowVersion = nil
		runtimeAgent.Executor.ShowTicks = nil
		runtimeAgent.Executor.ShowMemoryMap = nil
		runtimeAgent.Executor.SetMode = nil
		runtimeAgent.ExecutorConfigured = false
		return
	}
	runtimeAgent.Executor.ListFiles = runtime.Executor.ListFiles
	runtimeAgent.Executor.ReadFile = runtime.Executor.ReadFile
	runtimeAgent.Executor.WriteFile = runtime.Executor.WriteFile
	runtimeAgent.Executor.DeleteFile = runtime.Executor.DeleteFile
	runtimeAgent.Executor.StatFile = runtime.Executor.StatFile
	runtimeAgent.Executor.ShowHelp = runtime.Executor.ShowHelp
	runtimeAgent.Executor.ShowHistory = runtime.Executor.ShowHistory
	runtimeAgent.Executor.ShowVersion = runtime.Executor.ShowVersion
	runtimeAgent.Executor.ShowTicks = runtime.Executor.ShowTicks
	runtimeAgent.Executor.ShowMemoryMap = runtime.Executor.ShowMemoryMap
	runtimeAgent.Executor.SetMode = runtime.Executor.SetMode
	runtimeAgent.ExecutorConfigured = runtime.ExecutorConfigured
}

func ConfigureAgentRuntime() {
	runtimeAgent.Executor.ListFiles = agentListFiles
	runtimeAgent.Executor.ReadFile = agentReadFile
	runtimeAgent.Executor.WriteFile = nil
	runtimeAgent.Executor.DeleteFile = agentDeleteFile
	runtimeAgent.Executor.StatFile = agentStatFile
	runtimeAgent.Executor.ShowHelp = agentShowHelp
	runtimeAgent.Executor.ShowHistory = agentShowHistory
	runtimeAgent.Executor.ShowVersion = agentShowVersion
	runtimeAgent.Executor.ShowTicks = agentShowTicks
	runtimeAgent.Executor.ShowMemoryMap = agentShowMemoryMap
	runtimeAgent.Executor.SetMode = agentSetMode
	runtimeAgent.ExecutorConfigured = true
}

func NewAgentExecutor() agent.AllowedActionExecutor {
	var executor agent.AllowedActionExecutor
	executor.ListFiles = agentListFiles
	executor.ReadFile = agentReadFile
	executor.DeleteFile = agentDeleteFile
	executor.StatFile = agentStatFile
	executor.ShowHelp = agentShowHelp
	executor.ShowHistory = agentShowHistory
	executor.ShowVersion = agentShowVersion
	executor.ShowTicks = agentShowTicks
	executor.ShowMemoryMap = agentShowMemoryMap
	executor.SetMode = agentSetMode
	return executor
}

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

var targetLayout string

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

	if matchLiteral(cmdStart, cmdEnd, commandHistory) {
		printHistory()
		return
	}

	if matchLiteral(cmdStart, cmdEnd, commandHelp) {
		terminal.Print("Commands: help, history, clear, echo, ticks, uptime, mem, mmap, mmapmax, pfa, alloc, free, ls, write, cat, rm, stat, disk, fatinit, fatformat, fatinfo, fatls, fatcreate, fatread, layout, version, run, agent\n")
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
		if !printTicks() {
			terminal.Print("ticks: not wired yet\n")
		}
		return
	}

	if matchLiteral(cmdStart, cmdEnd, "uptime") {
		if getSyscallTicks == nil {
			terminal.Print("uptime: not wired yet\n")
			return
		}
		t := getSyscallTicks()
		secs := t / 100
		mins := secs / 60
		secs = secs % 60
		terminal.Print("up ")
		printUint(mins)
		terminal.Print("m ")
		printUint(secs)
		terminal.Print("s (")
		printUint(t)
		terminal.Print(" ticks via SYS_GETTICKS)\n")
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
		printMemoryMap()
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

	if matchLiteral(cmdStart, cmdEnd, "layout") {
		a1s, a1e, ok := nextArg(cmdEnd, end)
		if !ok {
			terminal.Print("current layout: ")
			terminal.Print(currentLayout)
			terminal.PutRune('\n')
			return
		}

		if matchLiteral(a1s, a1e, "us") {
			targetLayout = "us"
		} else if matchLiteral(a1s, a1e, "it") {
			targetLayout = "it"
		} else {
			terminal.Print("Usage: layout [us|it]\n")
			return
		}

		if switchLayoutFn == nil {
			terminal.Print("layout: switcher not wired\n")
			return
		}

		if !switchLayoutFn(targetLayout) {
			terminal.Print("layout: failed to switch\n")
			return
		}

		currentLayout = targetLayout
		terminal.Print("layout: switched to ")
		terminal.Print(targetLayout)
		terminal.PutRune('\n')
		return
	}

	if matchLiteral(cmdStart, cmdEnd, "version") {
		printVersion()
		return
	}

	if matchLiteral(cmdStart, cmdEnd, "run") {
		a1s, a1e, ok := nextArg(cmdEnd, end)
		if !ok {
			terminal.Print("Usage: run <program>\n")
			return
		}

		if runProgram == nil {
			terminal.Print("run: runner not wired\n")
			return
		}

		nameLen, ok := copyNameFromRange(a1s, a1e)
		if !ok {
			terminal.Print("run: invalid name\n")
			return
		}

		pid, ok := runProgram(&tmpName, nameLen)
		if !ok {
			terminal.Print("run: not found or no slot\n")
			return
		}

		terminal.Print("started pid=")
		printUint(uint64(pid))
		terminal.PutRune('\n')
		return
	}

	if matchLiteral(cmdStart, cmdEnd, "agent") {
		a1s, a1e, ok := nextArg(cmdEnd, end)
		if !ok {
			terminal.Print("Usage: agent <show|read|stat|delete|mode|help> [arg]\n")
			return
		}

		if matchLiteral(a1s, a1e, "show") {
			a2s, a2e, ok := nextArg(a1e, end)
			if !ok {
				terminal.Print("Usage: agent show <files|history|version|ticks|memorymap>\n")
				return
			}
			if matchLiteral(a2s, a2e, "files") {
				runAgentNoTarget(agent.ActionListFiles, agent.IntentListFiles, agent.RiskSafe)
				return
			} else if matchLiteral(a2s, a2e, "history") {
				runAgentNoTarget(agent.ActionShowHistory, agent.IntentShowHistory, agent.RiskSafe)
				return
			} else if matchLiteral(a2s, a2e, "version") {
				runAgentNoTarget(agent.ActionShowVersion, agent.IntentShowVersion, agent.RiskSafe)
				return
			} else if matchLiteral(a2s, a2e, "ticks") {
				runAgentNoTarget(agent.ActionShowTicks, agent.IntentShowTicks, agent.RiskSafe)
				return
			} else if matchLiteral(a2s, a2e, "memorymap") || matchLiteral(a2s, a2e, "memory_map") {
				runAgentNoTarget(agent.ActionShowMemoryMap, agent.IntentShowMemoryMap, agent.RiskSafe)
				return
			}
			terminal.Print("Usage: agent show <files|history|version|ticks|memorymap>\n")
			return
		} else if matchLiteral(a1s, a1e, "read") {
			a2s, a2e, ok := nextArg(a1e, end)
			if !ok {
				terminal.Print("Usage: agent read <name>\n")
				return
			}
			runAgentAction(agent.ActionReadFile, agent.IntentReadFile, agent.RiskSafe, a2s, a2e)
			return
		} else if matchLiteral(a1s, a1e, "delete") {
			a2s, a2e, ok := nextArg(a1e, end)
			if !ok {
				terminal.Print("Usage: agent delete <name> [confirm]\n")
				return
			}
			a3s, a3e, confirmed := nextArg(a2e, end)
			if confirmed && matchLiteral(a3s, a3e, "confirm") {
				runAgentAction(agent.ActionDeleteFile, agent.IntentDeleteFile, agent.RiskSafe, a2s, a2e)
				return
			}
			runAgentAction(agent.ActionDeleteFile, agent.IntentDeleteFile, agent.RiskRisky, a2s, a2e)
			return
		} else if matchLiteral(a1s, a1e, "stat") {
			a2s, a2e, ok := nextArg(a1e, end)
			if !ok {
				terminal.Print("Usage: agent stat <name>\n")
				return
			}
			runAgentAction(agent.ActionStatFile, agent.IntentStatFile, agent.RiskSafe, a2s, a2e)
			return
		} else if matchLiteral(a1s, a1e, "mode") {
			a2s, a2e, ok := nextArg(a1e, end)
			if ok {
				runAgentAction(agent.ActionSetMode, agent.IntentSetMode, agent.RiskSafe, a2s, a2e)
				return
			}
			runAgentNoTarget(agent.ActionSetMode, agent.IntentSetMode, agent.RiskSafe)
			return
		} else if matchLiteral(a1s, a1e, "help") {
			runAgentNoTarget(agent.ActionShowHelp, agent.IntentShowHelp, agent.RiskSafe)
			return
		}

		terminal.Print("Try with agent help to see available commands\n")
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

func runAgentNoTarget(kind agent.ActionKind, intent agent.IntentKind, risk agent.RiskLevel) {
	message := runtimeAgent.RunActionMessage(kind, intent, risk, nil, 0, nil)
	printAgentMessage(message)
	terminal.PutRune('\n')
}

func runAgentAction(kind agent.ActionKind, intent agent.IntentKind, risk agent.RiskLevel, targetStart, targetEnd int) {
	targetLen, ok := copyNameFromRange(targetStart, targetEnd)
	if !ok {
		terminal.Print("agent: invalid target\n")
		return
	}
	message := runtimeAgent.RunActionMessage(kind, intent, risk, &tmpName, targetLen, nil)
	printAgentMessage(message)
	terminal.PutRune('\n')
}

func printAgentMessage(message agent.MessageKind) {
	switch message {
	case agent.MessagePlannerFailed:
		terminal.Print("agent: planner failed")
	case agent.MessageValidationFailed:
		terminal.Print("validation_failed")
	case agent.MessagePlanHasNoActions:
		terminal.Print("agent: plan has no actions")
	case agent.MessagePlanHasTooManyActions:
		terminal.Print("agent: plan has too many actions")
	case agent.MessagePlanContainsUnsupportedAction:
		terminal.Print("agent: plan contains unsupported action")
	case agent.MessageActionRiskInvalid:
		terminal.Print("agent: action risk is invalid")
	case agent.MessageActionTargetInvalid:
		terminal.Print("agent: action target is invalid")
	case agent.MessageActionDataInvalid:
		terminal.Print("agent: action data is invalid")
	case agent.MessageConfirmationRequired:
		terminal.Print("agent: confirmation required")
	case agent.MessageExecutorNotConfigured:
		terminal.Print("agent: executor not configured")
	case agent.MessageUnsupportedAction:
		terminal.Print("agent: unsupported action")
	case agent.MessageActionUnavailable:
		terminal.Print("agent: action unavailable")
	case agent.MessageNoResult:
		terminal.Print("agent: no result")
	case agent.MessageCompletedPlan:
		terminal.Print("agent: completed plan")
	case agent.MessageOK:
		terminal.Print("ok")
	case agent.MessageFilesListed:
		terminal.Print("agent: files listed")
	case agent.MessageNoFiles:
		terminal.Print("agent: no files")
	case agent.MessageFileRead:
		terminal.Print("agent: file read")
	case agent.MessageFileStat:
		terminal.Print("agent: file stat")
	case agent.MessageMissingFile:
		terminal.Print("agent: missing file")
	case agent.MessageFileNotFound:
		terminal.Print("agent: file not found")
	case agent.MessageAgentHelp:
		terminal.Print("Agent commands:\n")
		terminal.Print("  agent show files    - Show files managed by the agent\n")
		terminal.Print("  agent show history  - Show command history stored by the agent\n")
		terminal.Print("  agent show version  - Show OS version through the agent\n")
		terminal.Print("  agent show ticks    - Show PIT ticks through the agent\n")
		terminal.Print("  agent show memorymap - Show memory map through the agent\n")
		terminal.Print("  agent read <name>   - Read a file through the agent\n")
		terminal.Print("  agent stat <name>   - Show file metadata through the agent\n")
		terminal.Print("  agent delete <name> confirm - Delete a file through the agent\n")
		terminal.Print("  agent mode [mode]   - Show or switch agent mode\n")
		terminal.Print("  agent help          - Show agent commands")
	case agent.MessageHistoryListed:
		terminal.Print("agent: history listed")
	case agent.MessageVersionShown:
		terminal.Print("agent: version shown")
	case agent.MessageTicksShown:
		terminal.Print("agent: ticks shown")
	case agent.MessageMemoryMapShown:
		terminal.Print("agent: memory map shown")
	case agent.MessageDeterministicMode:
		terminal.Print("agent: deterministic mode")
	case agent.MessageLLMModeNotConfigured:
		terminal.Print("agent: llm mode not configured")
	case agent.MessageUnsupportedMode:
		terminal.Print("agent: unsupported mode")
	default:
		return
	}
}

func printHistory() {
	startIdx := (historyHead - historyCount + maxHistory) % maxHistory
	for i := 0; i < historyCount; i++ {
		idx := (startIdx + i) % maxHistory
		printUint(uint64(i + 1))
		terminal.Print(" ")

		l := historyLen[idx]
		for k := 0; k < l; k++ {
			terminal.PutRune(rune(historyBuf[idx][k]))
		}
		terminal.PutRune('\n')
	}
}

func printVersion() {
	terminal.Print(osName + " " + osVersion)
	proof := uint64(0x0123456789ABCDEF)
	if (proof >> 32) != 0 {
		terminal.Print(" (64bit)")
	}
	terminal.PutRune('\n')
}

func printTicks() bool {
	if getTicks == nil {
		return false
	}
	printUint(getTicks())
	terminal.PutRune('\n')
	return true
}

func printMemoryMap() {
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
}

func agentListFiles(_ agent.Action, _ *agent.Context) agent.ActionResult {
	count := 0
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
		count++
	}
	if count == 0 {
		return agent.ActionResult{OK: true, Message: agent.MessageNoFiles}
	}
	return agent.ActionResult{OK: true, Message: agent.MessageFilesListed}
}

func agentReadFile(action agent.Action, _ *agent.Context) agent.ActionResult {
	if action.TargetLen <= 0 {
		return agent.ActionResult{OK: false, Message: agent.MessageMissingFile}
	}
	page, size, ok := fs.Lookup(&action.Target, action.TargetLen)
	if !ok {
		return agent.ActionResult{OK: false, Message: agent.MessageFileNotFound}
	}
	for i := uint64(0); i < size; i++ {
		b := *(*byte)(unsafe.Pointer(uintptr(page) + uintptr(i)))
		terminal.PutRune(rune(b))
	}
	terminal.PutRune('\n')
	return agent.ActionResult{OK: true, Message: agent.MessageFileRead}
}

func agentDeleteFile(action agent.Action, _ *agent.Context) agent.ActionResult {
	if action.TargetLen <= 0 {
		return agent.ActionResult{OK: false, Message: agent.MessageMissingFile}
	}
	if !fs.Remove(&action.Target, action.TargetLen) {
		return agent.ActionResult{OK: false, Message: agent.MessageFileNotFound}
	}
	return agent.ActionResult{OK: true, Message: agent.MessageOK}
}

func agentStatFile(action agent.Action, _ *agent.Context) agent.ActionResult {
	if action.TargetLen <= 0 {
		return agent.ActionResult{OK: false, Message: agent.MessageMissingFile}
	}
	page, size, ok := fs.Lookup(&action.Target, action.TargetLen)
	if !ok {
		return agent.ActionResult{OK: false, Message: agent.MessageFileNotFound}
	}
	terminal.Print("page=0x")
	printHexU64(page)
	terminal.Print(" size=")
	printUint(size)
	terminal.PutRune('\n')
	return agent.ActionResult{OK: true, Message: agent.MessageFileStat}
}

func agentShowHelp(_ agent.Action, _ *agent.Context) agent.ActionResult {
	return agent.ActionResult{OK: true, Message: agent.MessageAgentHelp}
}

func agentShowHistory(_ agent.Action, _ *agent.Context) agent.ActionResult {
	printHistory()
	return agent.ActionResult{OK: true, Message: agent.MessageHistoryListed}
}

func agentShowVersion(_ agent.Action, _ *agent.Context) agent.ActionResult {
	printVersion()
	return agent.ActionResult{OK: true, Message: agent.MessageVersionShown}
}

func agentShowTicks(_ agent.Action, _ *agent.Context) agent.ActionResult {
	if !printTicks() {
		return agent.ActionResult{OK: false, Message: agent.MessageActionUnavailable}
	}
	return agent.ActionResult{OK: true, Message: agent.MessageTicksShown}
}

func agentShowMemoryMap(_ agent.Action, _ *agent.Context) agent.ActionResult {
	printMemoryMap()
	return agent.ActionResult{OK: true, Message: agent.MessageMemoryMapShown}
}

func agentSetMode(action agent.Action, _ *agent.Context) agent.ActionResult {
	if action.TargetLen == 0 {
		return agent.ActionResult{OK: true, Message: agent.MessageDeterministicMode}
	}
	if actionTargetMatches(action, "deterministic") {
		return agent.ActionResult{OK: true, Message: agent.MessageDeterministicMode}
	}
	if actionTargetMatches(action, "llm") {
		return agent.ActionResult{OK: false, Message: agent.MessageLLMModeNotConfigured}
	}
	return agent.ActionResult{OK: false, Message: agent.MessageUnsupportedMode}
}

func actionTargetMatches(action agent.Action, literal string) bool {
	if action.TargetLen != len(literal) {
		return false
	}
	for i := 0; i < action.TargetLen; i++ {
		if action.Target[i] != literal[i] {
			return false
		}
	}
	return true
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
