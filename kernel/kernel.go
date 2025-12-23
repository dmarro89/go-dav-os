package kernel

import (
	"github.com/dmarro89/go-dav-os/keyboard"
	"github.com/dmarro89/go-dav-os/terminal"
)

func DebugChar(c byte)
func inb(port uint16) byte
func outb(port uint16, val byte)
func EnableInterrupts()
func DisableInterrupts()

func Main() {
	terminal.Init()
	terminal.Clear()

	InitIDT()
	TriggerInt80()
	terminal.Print("Back from int 0x80\n")

	for {
		terminal.Print("\n> ")

		var line [80]rune
		n := readLine(line[:])

		handleCommand(line[:n], n)
	}

}

func readLine(buf []rune) int {
	pos := 0

	for {
		r := keyboard.ReadKey()
		switch r {
		case '\b':
			if pos > 0 {
				pos--
				terminal.PutRune(r)
			}
		case '\n':
			terminal.PutRune(r)
			return pos
		default:
			if pos < len(buf) {
				buf[pos] = r
				pos++
				terminal.PutRune(r)
			}
		}
	}
}

func handleCommand(buf []rune, n int) {
	if n == 0 {
		return
	}

	if n == 4 &&
		buf[0] == 'h' &&
		buf[1] == 'e' &&
		buf[2] == 'l' &&
		buf[3] == 'p' {
		terminal.Print("Available commands:\n")
		terminal.Print("  help   - show this help\n")
		terminal.Print("  clear  - clear the screen\n")
		terminal.Print("  about  - info about DavideOS\n")
		return
	}

	if n == 5 &&
		buf[0] == 'c' &&
		buf[1] == 'l' &&
		buf[2] == 'e' &&
		buf[3] == 'a' &&
		buf[4] == 'r' {
		terminal.Clear()
		return
	}

	if n == 5 &&
		buf[0] == 'a' &&
		buf[1] == 'b' &&
		buf[2] == 'o' &&
		buf[3] == 'u' &&
		buf[4] == 't' {
		terminal.Print("DavideOS (Go) experimental kernel\n")
		terminal.Print("Powered by gccgo + custom runtime stubs.\n")
		return
	}

	terminal.Print("Unknown command: ")
	for i := 0; i < n; i++ {
		terminal.PutRune(buf[i])
	}
	terminal.Print("\nType 'help' for a list of commands.\n")
}

// utility functions for printing hex values
func hexNibble(n uint8) byte {
	if n < 10 {
		return '0' + n
	}
	return 'A' + (n - 10)
}

func printHex8(v uint8) {
	terminal.PutRune(rune(hexNibble(v >> 4)))
	terminal.PutRune(rune(hexNibble(v & 0xF)))
}

func printHex16(v uint16) {
	printHex8(uint8(v >> 8))
	printHex8(uint8(v))
}

func printHex32(v uint32) {
	printHex16(uint16(v >> 16))
	printHex16(uint16(v))
}
