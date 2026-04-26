//go:build !gccgo && !testing

package ata

func inb(port uint16) byte                     { return 0 }
func outb(port uint16, value byte)             {}
func insw(port uint16, addr *byte, count int)  {}
func outsw(port uint16, addr *byte, count int) {}
