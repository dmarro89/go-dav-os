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
	name := [8]byte{'F', 'I', 'L', 'E', ' ', ' ', ' ', ' '}
	ext := [3]byte{'T', 'X', 'T'}

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

func TestMaxCluster(t *testing.T) {
	// 100 sectors of data at 4 sectors per cluster = clusters 2..26
	if m := MaxCluster(200, 100, 4); m != 26 {
		t.Errorf("expected 26, got %d", m)
	}
	// 32-bit sector count (TotSec16 == 0): only the marker limit applies
	if m := MaxCluster(0, 100, 4); m != 0xFFF6 {
		t.Errorf("expected 0xFFF6, got 0x%X", m)
	}
	// a huge volume is capped below the bad-cluster marker
	if m := MaxCluster(0xFFFFFFFF, 100, 1); m != 0xFFF6 {
		t.Errorf("expected 0xFFF6, got 0x%X", m)
	}
}

func TestCheckEntry(t *testing.T) {
	const maxCluster = 0xFFF6
	tests := []struct {
		name         string
		cluster      uint16
		size         uint32
		max          uint16
		wantReadData bool
		wantOK       bool
	}{
		{"zero-length with cluster 0", 0, 0, maxCluster, false, true},
		{"zero-length with cluster 1", 1, 0, maxCluster, false, true},
		{"zero-length with marker cluster", 0xFFFF, 0, maxCluster, false, true},
		{"non-empty with cluster 0", 0, 10, maxCluster, false, false},
		{"non-empty with cluster 1", 1, 10, maxCluster, false, false},
		{"non-empty with cluster 2", 2, 10, maxCluster, true, true},
		{"non-empty with last data cluster", 0xFFF6, 512, maxCluster, true, true},
		{"non-empty with bad-cluster marker", 0xFFF7, 512, maxCluster, false, false},
		{"non-empty with end-of-chain marker", 0xFFFF, 512, maxCluster, false, false},
		{"non-empty at volume's last cluster", 26, 512, 26, true, true},
		{"non-empty past volume's last cluster", 27, 512, 26, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRead, gotOK := CheckEntry(tt.cluster, tt.size, tt.max)
			if gotRead != tt.wantReadData || gotOK != tt.wantOK {
				t.Fatalf("CheckEntry(%d, %d, %d) = (%v, %v), want (%v, %v)",
					tt.cluster, tt.size, tt.max, gotRead, gotOK, tt.wantReadData, tt.wantOK)
			}
		})
	}
}
