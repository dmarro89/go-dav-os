//go:build gccgo

package ata

func inb(port uint16) byte
func outb(port uint16, value byte)
func insw(port uint16, addr *byte, count int)
func outsw(port uint16, addr *byte, count int)
