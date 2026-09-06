package ata

import "testing"

func TestPrepareCommandFrameMaster(t *testing.T) {
	// LBA = 0x01234567:
	// byte 0 (LBALo): 0x67
	// byte 1 (LBAMid): 0x45
	// byte 2 (LBAHi): 0x23
	// top nibble (lba >> 24 & 0x0F): 0x01
	// Master headBase: 0xE0 | 0x01 = 0xE1
	lba := uint32(0x01234567)
	frame := PrepareCommandFrame(lba, 0, 1, CmdRead)

	if frame.DriveHead != 0xE1 {
		t.Fatalf("expected DriveHead 0xE1, got 0x%02X", frame.DriveHead)
	}
	if frame.SecCount != 1 {
		t.Fatalf("expected SecCount 1, got %d", frame.SecCount)
	}
	if frame.LBALo != 0x67 {
		t.Fatalf("expected LBALo 0x67, got 0x%02X", frame.LBALo)
	}
	if frame.LBAMid != 0x45 {
		t.Fatalf("expected LBAMid 0x45, got 0x%02X", frame.LBAMid)
	}
	if frame.LBAHi != 0x23 {
		t.Fatalf("expected LBAHi 0x23, got 0x%02X", frame.LBAHi)
	}
	if frame.Command != CmdRead {
		t.Fatalf("expected Command 0x%02X, got 0x%02X", CmdRead, frame.Command)
	}
}

func TestPrepareCommandFrameSlaveWrite(t *testing.T) {
	// LBA = 0x0A000000:
	// top nibble: 0x0A
	// Slave headBase: 0xF0 | 0x0A = 0xFA
	lba := uint32(0x0A000000)
	frame := PrepareCommandFrame(lba, 1, 8, CmdWrite)

	if frame.DriveHead != 0xFA {
		t.Fatalf("expected DriveHead 0xFA, got 0x%02X", frame.DriveHead)
	}
	if frame.SecCount != 8 {
		t.Fatalf("expected SecCount 8, got %d", frame.SecCount)
	}
	if frame.LBALo != 0 || frame.LBAMid != 0 || frame.LBAHi != 0 {
		t.Fatalf("expected LBA bytes 0, got %02X %02X %02X", frame.LBALo, frame.LBAMid, frame.LBAHi)
	}
	if frame.Command != CmdWrite {
		t.Fatalf("expected Command 0x%02X, got 0x%02X", CmdWrite, frame.Command)
	}
}

func TestByteOffsetToLBA(t *testing.T) {
	tests := []struct {
		offset       uint64
		expectedLBA  uint32
		expectedSecO uint16
	}{
		{offset: 0, expectedLBA: 0, expectedSecO: 0},
		{offset: 511, expectedLBA: 0, expectedSecO: 511},
		{offset: 512, expectedLBA: 1, expectedSecO: 0},
		{offset: 1025, expectedLBA: 2, expectedSecO: 1},
		{offset: 20 * 1024 * 1024, expectedLBA: 40960, expectedSecO: 0},
	}

	for _, tt := range tests {
		lba, secOff := ByteOffsetToLBA(tt.offset)
		if lba != tt.expectedLBA || secOff != tt.expectedSecO {
			t.Errorf("ByteOffsetToLBA(%d) = (%d, %d), expected (%d, %d)",
				tt.offset, lba, secOff, tt.expectedLBA, tt.expectedSecO)
		}
		// Invariant: LBAToByteOffset reverses ByteOffsetToLBA
		reconstructed := LBAToByteOffset(lba, secOff)
		if reconstructed != tt.offset {
			t.Errorf("LBAToByteOffset(%d, %d) = %d, expected %d", lba, secOff, reconstructed, tt.offset)
		}
	}
}

func TestSectorsNeeded(t *testing.T) {
	tests := []struct {
		offset   uint64
		length   uint64
		expected uint32
	}{
		{offset: 0, length: 0, expected: 0},
		{offset: 0, length: 1, expected: 1},
		{offset: 0, length: 512, expected: 1},
		{offset: 0, length: 513, expected: 2},
		{offset: 100, length: 412, expected: 1}, // byte 100..511 fits in sector 0
		{offset: 100, length: 413, expected: 2}, // byte 100..512 crosses into sector 1
		{offset: 511, length: 2, expected: 2},   // byte 511..512 crosses sector boundary
		{offset: 1024, length: 1024, expected: 2},
	}

	for _, tt := range tests {
		got := SectorsNeeded(tt.offset, tt.length)
		if got != tt.expected {
			t.Errorf("SectorsNeeded(%d, %d) = %d, expected %d", tt.offset, tt.length, got, tt.expected)
		}
	}
}

func TestStatusBits(t *testing.T) {
	// BSY: 0x80
	if !IsStatusBusy(0x80) || !IsStatusBusy(0xFF) {
		t.Errorf("expected IsStatusBusy to be true when 0x80 is set")
	}
	if IsStatusBusy(0x7F) || IsStatusBusy(0x00) {
		t.Errorf("expected IsStatusBusy to be false when 0x80 is cleared")
	}

	// DRQ: 0x08
	if !IsStatusDRQ(0x08) || !IsStatusDRQ(0x0F) {
		t.Errorf("expected IsStatusDRQ to be true when 0x08 is set")
	}
	if IsStatusDRQ(0xF7) || IsStatusDRQ(0x00) {
		t.Errorf("expected IsStatusDRQ to be false when 0x08 is cleared")
	}

	// ERR / DF: 0x01 or 0x20
	if !IsStatusError(0x01) {
		t.Errorf("expected IsStatusError(0x01) to be true")
	}
	if !IsStatusError(0x20) {
		t.Errorf("expected IsStatusError(0x20) to be true")
	}
	if IsStatusError(0x08) || IsStatusError(0x00) {
		t.Errorf("expected IsStatusError(0x08/0x00) to be false")
	}
}
