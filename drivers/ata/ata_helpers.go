package ata

// LBARegs holds the values for the LBA registers.
type LBARegs struct {
	DriveHead byte
	LBALo     byte
	LBAMid    byte
	LBAHi     byte
}

// LBA28ToRegs converts a 28-bit LBA address to ATA registers format.
func LBA28ToRegs(lba uint32) LBARegs {
	return LBARegs{
		DriveHead: 0xE0 | byte((lba>>24)&0x0F),
		LBALo:     byte(lba),
		LBAMid:    byte(lba >> 8),
		LBAHi:     byte(lba >> 16),
	}
}
