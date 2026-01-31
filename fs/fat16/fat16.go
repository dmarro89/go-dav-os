package fat16

import (
	"github.com/dmarro89/go-dav-os/drivers/ata"
	"github.com/dmarro89/go-dav-os/terminal"
)

var (
	BytesPerSec uint16
	SecPerClust uint8
	ReservedSec uint16
	NumFATs     uint8
	RootEntCnt  uint16
	TotSec16    uint16
	FatSz16     uint16

	// Computed Offsets (LBA)
	fatStart    uint32
	rootStart   uint32
	dataStart   uint32
	rootSectors uint32

	initialized bool

	// Global buffer to avoid runtime.newobject (heap allocation)
	fatBuf [512]byte
)

const (
	DirEntrySize = 32
)

// Init reads the MBR/BPB from sector 0 and calculates offsets
func Init() bool {
	if !ata.ReadSector(0, &fatBuf) {
		terminal.Print("FAT16: Read Error\n")
		return false
	}

	// Check signature 0x55 0xAA at 510
	if fatBuf[510] != 0x55 || fatBuf[511] != 0xAA {
		terminal.Print("FAT16: Invalid Signature. Run 'fat_format' first.\n")
		initialized = false
		return false
	}

	BytesPerSec = uint16(fatBuf[11]) | uint16(fatBuf[12])<<8
	SecPerClust = fatBuf[13]
	ReservedSec = uint16(fatBuf[14]) | uint16(fatBuf[15])<<8
	NumFATs = fatBuf[16]
	RootEntCnt = uint16(fatBuf[17]) | uint16(fatBuf[18])<<8
	TotSec16 = uint16(fatBuf[19]) | uint16(fatBuf[20])<<8
	FatSz16 = uint16(fatBuf[22]) | uint16(fatBuf[23])<<8

	if BytesPerSec != 512 {
		terminal.Print("FAT16 Error: BytesPerSec != 512\n")
		return false
	}

	fatStart = uint32(ReservedSec)
	rootStart = fatStart + (uint32(NumFATs) * uint32(FatSz16))

	// Root dir size in sectors
	rootSectors = (uint32(RootEntCnt)*32 + 511) / 512
	dataStart = rootStart + rootSectors

	initialized = true
	return true
}

// Format writes a minimal FAT16 BPB to sector 0
func Format() bool {
	// Clear buffer
	for i := 0; i < 512; i++ {
		fatBuf[i] = 0
	}

	// Jump
	fatBuf[0], fatBuf[1], fatBuf[2] = 0xEB, 0x3C, 0x90
	// OEM
	oem := "MSWIN4.1"
	for i := 0; i < len(oem); i++ {
		fatBuf[3+i] = oem[i]
	}

	// BPB
	// BytesPerSec = 512
	fatBuf[11], fatBuf[12] = 0x00, 0x02
	// SecPerClust = 1
	fatBuf[13] = 1
	// ReservedSec = 1 (Boot sector)
	fatBuf[14], fatBuf[15] = 1, 0
	// NumFATs = 2
	fatBuf[16] = 2
	// RootEntCnt = 512 (Standard)
	fatBuf[17], fatBuf[18] = 0x00, 0x02
	// TotSec16 = 40960 (20MB / 512)
	fatBuf[19], fatBuf[20] = 0x00, 0xA0
	// Media = F8 (Fixed)
	fatBuf[21] = 0xF8
	// FatSz16 = 160 sectors
	fatBuf[22], fatBuf[23] = 160, 0

	// Signature
	fatBuf[510] = 0x55
	fatBuf[511] = 0xAA

	if !ata.WriteSector(0, &fatBuf) {
		return false
	}

	// Zero out FAT1 start and RootDir start
	for i := 0; i < 512; i++ {
		fatBuf[i] = 0
	}

	// FAT1 starts at ReservedSec (1)
	ata.WriteSector(uint32(1), &fatBuf)

	// Root starts at 1 + 2*160 = 321
	ata.WriteSector(321, &fatBuf)

	return true
}

func Info() {
	if !initialized {
		terminal.Print("FAT16: Not initialized\n")
		return
	}
	terminal.Print("FAT16 Layout:\n")
	terminal.Print("  Reverved Sec: ")
	printU16(ReservedSec)
	terminal.Print("\n  FAT Start: ")
	printU32(fatStart)
	terminal.Print("\n  Root Start: ")
	printU32(rootStart)
	terminal.Print("\n  Data Start: ")
	printU32(dataStart)
	terminal.Print("\n")
}

func printU16(v uint16) {
	printU32(uint32(v))
}

func printU32(v uint32) {
	terminal.Print("0x")
	hex := "0123456789ABCDEF"
	for i := 28; i >= 0; i -= 4 {
		terminal.PutRune(rune(hex[(v>>uint(i))&0xF]))
	}
}
