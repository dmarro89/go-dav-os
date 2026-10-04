package fat16

// ClusterToSector converts a FAT16 cluster number to an LBA sector
func ClusterToSector(dataStart uint32, secPerClust uint8, cluster uint16) uint32 {
	if cluster < 2 {
		return dataStart // Fallback for invalid cluster
	}
	return dataStart + uint32(cluster-2)*uint32(secPerClust)
}

// MatchName checks if a directory entry matches the requested 8.3 filename
func MatchName(entry []byte, name *[8]byte, ext *[3]byte) bool {
	if len(entry) < 11 {
		return false
	}
	for i := 0; i < 8; i++ {
		if entry[i] != name[i] {
			return false
		}
	}
	for i := 0; i < 3; i++ {
		if entry[8+i] != ext[i] {
			return false
		}
	}
	return true
}

// ParseEntry reads the cluster and size from a 32-byte FAT16 directory entry
func ParseEntry(entry []byte) (cluster uint16, size uint32) {
	if len(entry) < 32 {
		return 0, 0
	}
	cluster = uint16(entry[26]) | uint16(entry[27])<<8
	size = uint32(entry[28]) | uint32(entry[29])<<8 | uint32(entry[30])<<16 | uint32(entry[31])<<24
	return cluster, size
}
