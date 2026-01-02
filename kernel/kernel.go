package kernel

import (
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

func Main(multibootInfoAddr uint32) {
	DisableInterrupts()
	terminal.Init()
	terminal.Clear()

	InitIDT()

	PICRemap(0x20, 0x28)
	PICSetMask(0xFC, 0xFF)

	PITInit(100)

	shell.SetTickProvider(GetTicks)

	mem.InitMultiboot(multibootInfoAddr)
	mem.InitPFA()

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
