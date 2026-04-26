//go:build gccgo

package terminal

func outb(port uint16, value byte)
func debugChar(c byte)
