package ata

import "unsafe"

func inb(port uint16) byte
func outb(port uint16, value byte)
func insw(port uint16, addr *byte, count int)
func outsw(port uint16, addr *byte, count int)

const (
	Data      uint16 = 0x1F0
	ErrFeat   uint16 = 0x1F1
	SecCount  uint16 = 0x1F2
	LBALo     uint16 = 0x1F3
	LBAMid    uint16 = 0x1F4
	LBAHi     uint16 = 0x1F5
	DriveHead uint16 = 0x1F6
	StatusCmd uint16 = 0x1F7

	CmdRead  = 0x20
	CmdWrite = 0x30
	CmdFlush = 0xE7
)

// Timeout constant for ATA operations (iterations)
const ataTimeout = 100000

func waitBusy() bool {
	for i := 0; i < ataTimeout; i++ {
		status := inb(StatusCmd)
		if (status & 0x80) == 0 {
			return true
		}
	}
	return false // Timeout
}

func waitDRQ() bool {
	for i := 0; i < ataTimeout; i++ {
		status := inb(StatusCmd)
		if (status & 0x01) != 0 {
			return false // ERR
		}
		if (status & 0x08) != 0 {
			return true // DRQ ready
		}
	}
	return false // Timeout
}

func ReadSector(lba uint32, buf *[512]byte) bool {
	if !waitBusy() {
		return false
	}

	outb(DriveHead, 0xE0|byte((lba>>24)&0x0F))
	outb(SecCount, 1)
	outb(LBALo, byte(lba))
	outb(LBAMid, byte(lba>>8))
	outb(LBAHi, byte(lba>>16))
	outb(StatusCmd, CmdRead)

	if !waitDRQ() {
		return false
	}

	insw(Data, (*byte)(unsafe.Pointer(&buf[0])), 256)
	return true
}

func WriteSector(lba uint32, data *[512]byte) bool {
	if !waitBusy() {
		return false
	}

	outb(DriveHead, 0xE0|byte((lba>>24)&0x0F))
	outb(SecCount, 1)
	outb(LBALo, byte(lba))
	outb(LBAMid, byte(lba>>8))
	outb(LBAHi, byte(lba>>16))
	outb(StatusCmd, CmdWrite)

	if !waitDRQ() {
		return false
	}

	outsw(Data, (*byte)(unsafe.Pointer(&data[0])), 256)

	// Flush Cache
	outb(StatusCmd, CmdFlush)
	if !waitBusy() {
		return false
	}

	return true
}
