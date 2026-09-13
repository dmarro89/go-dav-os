//go:build !testing

package ata

import "unsafe"

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

// waitBusy waits for the ATA drive to clear its busy flag, returning false on timeout.
func waitBusy() bool {
	for i := 0; i < ataTimeout; i++ {
		status := inb(StatusCmd)
		if (status & 0x80) == 0 {
			return true
		}
	}
	return false // Timeout
}

// waitDRQ waits for the ATA drive to set its Data Request (DRQ) flag or Error (ERR) flag, returning false if an error or timeout occurs.
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

// ReadSector reads a single 512-byte sector from the given 28-bit LBA address into buf.
func ReadSector(lba uint32, buf *[512]byte) bool {
	if !waitBusy() {
		return false
	}

	regs := LBA28ToRegs(lba)
	outb(DriveHead, regs.DriveHead)
	outb(SecCount, 1)
	outb(LBALo, regs.LBALo)
	outb(LBAMid, regs.LBAMid)
	outb(LBAHi, regs.LBAHi)
	outb(StatusCmd, CmdRead)

	if !waitDRQ() {
		return false
	}

	insw(Data, (*byte)(unsafe.Pointer(&buf[0])), 256)
	return true
}

// WriteSector writes a single 512-byte sector from data to the given 28-bit LBA address.
func WriteSector(lba uint32, data *[512]byte) bool {
	if !waitBusy() {
		return false
	}

	regs := LBA28ToRegs(lba)
	outb(DriveHead, regs.DriveHead)
	outb(SecCount, 1)
	outb(LBALo, regs.LBALo)
	outb(LBAMid, regs.LBAMid)
	outb(LBAHi, regs.LBAHi)
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
