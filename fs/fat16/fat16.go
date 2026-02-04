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

	// Prepare zero buffer for clearing sectors
	for i := 0; i < 512; i++ {
		fatBuf[i] = 0
	}

	// FatSz16 = 160 sectors per FAT
	// NumFATs = 2
	// ReservedSec = 1
	const fatSz16 uint32 = 160
	const reservedSec uint32 = 1

	// Zero out all sectors of FAT1 (sectors 1 to 160)
	for sec := uint32(0); sec < fatSz16; sec++ {
		if !ata.WriteSector(reservedSec+sec, &fatBuf) {
			return false
		}
	}

	// Zero out all sectors of FAT2 (sectors 161 to 320)
	for sec := uint32(0); sec < fatSz16; sec++ {
		if !ata.WriteSector(reservedSec+fatSz16+sec, &fatBuf) {
			return false
		}
	}

	// Root directory starts at sector 321 (1 + 160 + 160)
	// Root has 512 entries * 32 bytes = 16384 bytes = 32 sectors
	const rootStart uint32 = reservedSec + fatSz16*2 // = 321
	const rootSectors uint32 = 32

	// Zero out all root directory sectors
	for sec := uint32(0); sec < rootSectors; sec++ {
		if !ata.WriteSector(rootStart+sec, &fatBuf) {
			return false
		}
	}

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

// ListDir lists files in the root directory
func ListDir() {
	if !initialized {
		terminal.Print("FAT16: Not initialized\n")
		return
	}

	terminal.Print("Root Directory:\n")
	entriesPerSector := 512 / DirEntrySize // 16

	for sec := uint32(0); sec < rootSectors; sec++ {
		if !ata.ReadSector(rootStart+sec, &fatBuf) {
			terminal.Print("FAT16: Read error\n")
			return
		}

		for i := 0; i < entriesPerSector; i++ {
			off := i * DirEntrySize
			firstByte := fatBuf[off]

			if firstByte == 0x00 {
				return // No more entries
			}
			if firstByte == 0xE5 {
				continue // Deleted entry
			}
			if fatBuf[off+11]&0x08 != 0 {
				continue // Volume label
			}

			// Print filename (8 chars) + ext (3 chars)
			terminal.Print("  ")
			for j := 0; j < 8; j++ {
				c := fatBuf[off+j]
				if c != ' ' {
					terminal.PutRune(rune(c))
				}
			}
			if fatBuf[off+8] != ' ' {
				terminal.PutRune('.')
				for j := 8; j < 11; j++ {
					c := fatBuf[off+j]
					if c != ' ' {
						terminal.PutRune(rune(c))
					}
				}
			}

			// Size (bytes 28-31, little endian)
			size := uint32(fatBuf[off+28]) | uint32(fatBuf[off+29])<<8 |
				uint32(fatBuf[off+30])<<16 | uint32(fatBuf[off+31])<<24
			terminal.Print("  ")
			printU32(size)
			terminal.Print(" bytes\n")
		}
	}
}

// CreateFile creates a file in the root directory
func CreateFile(name *[8]byte, ext *[3]byte, data *[512]byte, dataLen uint32) bool {
	if !initialized {
		terminal.Print("FAT16: Not initialized\n")
		return false
	}

	// Check if file with same name already exists
	for sec := uint32(0); sec < rootSectors; sec++ {
		if !ata.ReadSector(rootStart+sec, &fatBuf) {
			return false
		}

		for i := 0; i < 16; i++ {
			off := i * DirEntrySize
			firstByte := fatBuf[off]

			if firstByte == 0x00 {
				// End of directory, no duplicate found
				break
			}
			if firstByte == 0xE5 {
				continue // Deleted entry
			}
			if fatBuf[off+11]&0x08 != 0 {
				continue // Volume label
			}

			// Compare name and extension
			match := true
			for j := 0; j < 8; j++ {
				if fatBuf[off+j] != name[j] {
					match = false
					break
				}
			}
			if match {
				for j := 0; j < 3; j++ {
					if fatBuf[off+8+j] != ext[j] {
						match = false
						break
					}
				}
			}
			if match {
				terminal.Print("FAT16: File already exists\n")
				return false
			}
		}
	}

	// Find free cluster in FAT
	cluster := findFreeCluster()
	if cluster == 0 {
		terminal.Print("FAT16: No free clusters\n")
		return false
	}

	// Mark cluster as end-of-chain in FAT
	if !setFATEntry(cluster, 0xFFFF) {
		terminal.Print("FAT16: FAT write error\n")
		return false
	}

	// Find free directory entry
	entryFound := false
	var dirSec uint32
	var dirOff int

	for sec := uint32(0); sec < rootSectors && !entryFound; sec++ {
		if !ata.ReadSector(rootStart+sec, &fatBuf) {
			return false
		}

		for i := 0; i < 16; i++ {
			off := i * DirEntrySize
			firstByte := fatBuf[off]
			if firstByte == 0x00 || firstByte == 0xE5 {
				dirSec = sec
				dirOff = off
				entryFound = true
				break
			}
		}
	}

	if !entryFound {
		terminal.Print("FAT16: Root directory full\n")
		return false
	}

	// Re-read sector for modification
	if !ata.ReadSector(rootStart+dirSec, &fatBuf) {
		return false
	}

	// Write directory entry
	// Filename (8 bytes, space padded)
	for i := 0; i < 8; i++ {
		fatBuf[dirOff+i] = name[i]
	}
	// Extension (3 bytes, space padded)
	for i := 0; i < 3; i++ {
		fatBuf[dirOff+8+i] = ext[i]
	}
	// Attributes (0x00 = normal file)
	fatBuf[dirOff+11] = 0x00
	// Reserved bytes
	for i := 12; i < 26; i++ {
		fatBuf[dirOff+i] = 0
	}
	// First cluster (bytes 26-27, little endian)
	fatBuf[dirOff+26] = byte(cluster & 0xFF)
	fatBuf[dirOff+27] = byte((cluster >> 8) & 0xFF)
	// File size (bytes 28-31, little endian)
	fatBuf[dirOff+28] = byte(dataLen & 0xFF)
	fatBuf[dirOff+29] = byte((dataLen >> 8) & 0xFF)
	fatBuf[dirOff+30] = byte((dataLen >> 16) & 0xFF)
	fatBuf[dirOff+31] = byte((dataLen >> 24) & 0xFF)

	if !ata.WriteSector(rootStart+dirSec, &fatBuf) {
		return false
	}

	// Write data to cluster
	dataSector := clusterToSector(cluster)
	// Copy data to fatBuf
	for i := 0; i < 512; i++ {
		if uint32(i) < dataLen {
			fatBuf[i] = data[i]
		} else {
			fatBuf[i] = 0
		}
	}
	if !ata.WriteSector(dataSector, &fatBuf) {
		return false
	}

	return true
}

// ReadFile reads a file by name into the provided buffer
func ReadFile(name *[8]byte, ext *[3]byte, outBuf *[512]byte) (uint32, bool) {
	if !initialized {
		return 0, false
	}

	// Find file in root directory
	for sec := uint32(0); sec < rootSectors; sec++ {
		if !ata.ReadSector(rootStart+sec, &fatBuf) {
			return 0, false
		}

		for i := 0; i < 16; i++ {
			off := i * DirEntrySize
			firstByte := fatBuf[off]

			if firstByte == 0x00 {
				return 0, false // End of directory
			}
			if firstByte == 0xE5 {
				continue
			}

			// Compare name
			match := true
			for j := 0; j < 8; j++ {
				if fatBuf[off+j] != name[j] {
					match = false
					break
				}
			}
			if match {
				for j := 0; j < 3; j++ {
					if fatBuf[off+8+j] != ext[j] {
						match = false
						break
					}
				}
			}

			if match {
				// Found! Get cluster and size
				cluster := uint16(fatBuf[off+26]) | uint16(fatBuf[off+27])<<8
				size := uint32(fatBuf[off+28]) | uint32(fatBuf[off+29])<<8 |
					uint32(fatBuf[off+30])<<16 | uint32(fatBuf[off+31])<<24

				// Read data from cluster
				dataSector := clusterToSector(cluster)
				if !ata.ReadSector(dataSector, outBuf) {
					return 0, false
				}
				return size, true
			}
		}
	}

	return 0, false
}

// findFreeCluster finds a free cluster in the FAT (returns 0 if none)
func findFreeCluster() uint16 {
	// FAT16: each entry is 2 bytes
	// Clusters 0 and 1 are reserved, start at 2
	entriesPerSector := 256 // 512 / 2

	for sec := uint32(0); sec < uint32(FatSz16); sec++ {
		if !ata.ReadSector(fatStart+sec, &fatBuf) {
			return 0
		}

		for i := 0; i < entriesPerSector; i++ {
			cluster := sec*256 + uint32(i)
			if cluster < 2 {
				continue // Reserved
			}

			entry := uint16(fatBuf[i*2]) | uint16(fatBuf[i*2+1])<<8
			if entry == 0x0000 {
				return uint16(cluster)
			}
		}
	}
	return 0
}

// setFATEntry sets a FAT entry value
func setFATEntry(cluster uint16, value uint16) bool {
	// Calculate which sector and offset
	fatOffset := uint32(cluster) * 2
	sec := fatOffset / 512
	off := fatOffset % 512

	if !ata.ReadSector(fatStart+sec, &fatBuf) {
		return false
	}

	fatBuf[off] = byte(value & 0xFF)
	fatBuf[off+1] = byte((value >> 8) & 0xFF)

	// Write to both FATs
	if !ata.WriteSector(fatStart+sec, &fatBuf) {
		return false
	}
	// FAT2
	ata.WriteSector(fatStart+uint32(FatSz16)+sec, &fatBuf)

	return true
}

// clusterToSector converts a cluster number to LBA sector
func clusterToSector(cluster uint16) uint32 {
	return dataStart + uint32(cluster-2)*uint32(SecPerClust)
}
