package ata

import "testing"

func TestLBA28ToRegs(t *testing.T) {
	tests := []struct {
		name     string
		lba      uint32
		expected LBARegs
	}{
		{
			name: "zero lba",
			lba:  0,
			expected: LBARegs{
				DriveHead: 0xE0,
				LBALo:     0,
				LBAMid:    0,
				LBAHi:     0,
			},
		},
		{
			name: "max 28-bit lba",
			lba:  0x0FFFFFFF,
			expected: LBARegs{
				DriveHead: 0xEF,
				LBALo:     0xFF,
				LBAMid:    0xFF,
				LBAHi:     0xFF,
			},
		},
		{
			name: "random lba",
			lba:  0x01234567,
			expected: LBARegs{
				DriveHead: 0xE1,
				LBALo:     0x67,
				LBAMid:    0x45,
				LBAHi:     0x23,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LBA28ToRegs(tt.lba)
			if got != tt.expected {
				t.Errorf("LBA28ToRegs(%#x) = %+v, want %+v", tt.lba, got, tt.expected)
			}
		})
	}
}
