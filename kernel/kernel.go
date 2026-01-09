package kernel

import (
	"github.com/dmarro89/go-dav-os/fs"
	"github.com/dmarro89/go-dav-os/keyboard"
	"github.com/dmarro89/go-dav-os/mem"
	"github.com/dmarro89/go-dav-os/shell"
	"github.com/dmarro89/go-dav-os/terminal"
)

func DebugChar(c byte)
func inb(port uint16) byte
func outb(port uint16, val byte)
func EnableInterrupts()
func DisableInterrupts()
func Halt()

var syscallMsg = [...]byte{
	'h','e','l','l','o',' ','v','i','a',' ','s','y','s','c','a','l','l',' ','w','r','i','t','e','\n',
}

func SyscallTest() {
	TriggerSysWrite(&syscallMsg[0], uint32(len(syscallMsg)))
}

func Main(multibootInfoAddr uint32) {
	DisableInterrupts()
	terminal.Init()
	terminal.Clear()

	InitIDT()

	SyscallTest()

	PICRemap(0x20, 0x28)
	PICSetMask(0xFC, 0xFF)

	PITInit(100)

	shell.SetTickProvider(GetTicks)

	mem.InitMultiboot(multibootInfoAddr)
	mem.InitPFA()

	fs.Init()

	EnableInterrupts()
	shell.Init()

	for {
		DisableInterrupts()
		r, ok := keyboard.TryRead()
		EnableInterrupts()
		if !ok {
			Halt()
			continue
		}
		shell.FeedRune(r)
	}
}
