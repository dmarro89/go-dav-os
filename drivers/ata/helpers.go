package ata

// SectorSize is the standard ATA logical sector size in bytes.
const SectorSize = 512

const (
	CmdRead  = 0x20
	CmdWrite = 0x30
	CmdFlush = 0xE7
)

// CommandFrame represents the I/O register values prepared for an ATA command.
type CommandFrame struct {
	DriveHead byte
	SecCount  byte
	LBALo     byte
	LBAMid    byte
	LBAHi     byte
	Command   byte
}

// PrepareCommandFrame constructs the register state for 28-bit LBA ATA command.
// drive: 0 for master (0xE0 base), 1 for slave (0xF0 base).
// secCount: number of sectors (e.g. 1).
// command: ATA command byte (e.g. CmdRead, CmdWrite).
func PrepareCommandFrame(lba uint32, drive byte, secCount byte, command byte) CommandFrame {
	headBase := byte(0xE0)
	if (drive & 1) != 0 {
		headBase = 0xF0
	}
	return CommandFrame{
		DriveHead: headBase | byte((lba>>24)&0x0F),
		SecCount:  secCount,
		LBALo:     byte(lba),
		LBAMid:    byte(lba >> 8),
		LBAHi:     byte(lba >> 16),
		Command:   command,
	}
}

// ByteOffsetToLBA converts a 64-bit byte offset to sector LBA and sector offset.
func ByteOffsetToLBA(offset uint64) (lba uint32, sectorOffset uint16) {
	lba = uint32(offset / SectorSize)
	sectorOffset = uint16(offset % SectorSize)
	return lba, sectorOffset
}

// LBAToByteOffset converts an LBA and within-sector offset to a 64-bit byte offset.
func LBAToByteOffset(lba uint32, sectorOffset uint16) uint64 {
	return uint64(lba)*SectorSize + uint64(sectorOffset)
}

// SectorsNeeded calculates how many 512-byte sectors are needed to store length bytes starting from offset.
func SectorsNeeded(offset uint64, length uint64) uint32 {
	if length == 0 {
		return 0
	}
	startSector := offset / SectorSize
	endSector := (offset + length - 1) / SectorSize
	return uint32(endSector - startSector + 1)
}

// IsStatusBusy returns true if the BSY (Busy) bit (0x80) is set.
func IsStatusBusy(status byte) bool {
	return (status & 0x80) != 0
}

// IsStatusDRQ returns true if the DRQ (Data Request) bit (0x08) is set.
func IsStatusDRQ(status byte) bool {
	return (status & 0x08) != 0
}

// IsStatusError returns true if the ERR bit (0x01) or DF (Drive Fault) bit (0x20) is set.
func IsStatusError(status byte) bool {
	return (status & 0x21) != 0
}
