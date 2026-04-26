//go:build !gccgo && !testing

package terminal

func outb(port uint16, value byte) {}
func debugChar(c byte)             {}
