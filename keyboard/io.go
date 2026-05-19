//go:build !testing

package keyboard

func inb(port uint16) byte
func outb(port uint16, value byte)
