package fat16

import (
	"testing"
)

func TestClusterToSector(t *testing.T) {
	dataStart := uint32(100)
	secPerClust := uint8(4)

	// cluster 2 should be at dataStart (cluster 0 and 1 are reserved)
	if s := ClusterToSector(dataStart, secPerClust, 2); s != 100 {
		t.Errorf("expected 100, got %d", s)
	}

	// cluster 3 should be at dataStart + 4
	if s := ClusterToSector(dataStart, secPerClust, 3); s != 104 {
		t.Errorf("expected 104, got %d", s)
	}
}

func TestMatchName(t *testing.T) {
	name := [8]byte{'F','I','L','E',' ',' ',' ',' '}
	ext := [3]byte{'T','X','T'}

	entry := make([]byte, 32)
	copy(entry[0:8], name[:])
	copy(entry[8:11], ext[:])

	if !MatchName(entry, &name, &ext) {
		t.Errorf("expected exact match to succeed")
	}

	badExt := ext
	badExt[0] = 'B'
	if MatchName(entry, &name, &badExt) {
		t.Errorf("expected mismatch on ext to fail")
	}

	badName := name
	badName[0] = 'X'
	if MatchName(entry, &badName, &ext) {
		t.Errorf("expected mismatch on name to fail")
	}
}

func TestParseEntry(t *testing.T) {
	entry := make([]byte, 32)
	// Cluster = 0x1234 (offset 26)
	entry[26] = 0x34
	entry[27] = 0x12
	// Size = 0xDEADBEEF (offset 28)
	entry[28] = 0xEF
	entry[29] = 0xBE
	entry[30] = 0xAD
	entry[31] = 0xDE

	cluster, size := ParseEntry(entry)
	if cluster != 0x1234 {
		t.Errorf("expected cluster 0x1234, got 0x%X", cluster)
	}
	if size != 0xDEADBEEF {
		t.Errorf("expected size 0xDEADBEEF, got 0x%X", size)
	}
}
