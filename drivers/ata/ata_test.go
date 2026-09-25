package ata

import (
	"testing"
)

func TestWaitBusy(t *testing.T) {
	t.Run("busy clears immediately", func(t *testing.T) {
		ResetMocks()
		defer ResetMocks()
		calls := 0
		MockInb = func(port uint16) byte {
			if port != StatusCmd {
				t.Errorf("expected read from StatusCmd, got %x", port)
			}
			calls++
			return 0x00 // Not busy
		}
		if !waitBusy() {
			t.Error("expected waitBusy to return true")
		}
		if calls != 1 {
			t.Errorf("expected 1 call, got %d", calls)
		}
	})

	t.Run("busy times out", func(t *testing.T) {
		ResetMocks()
		defer ResetMocks()
		calls := 0
		MockInb = func(port uint16) byte {
			calls++
			return 0x80 // Always busy
		}
		if waitBusy() {
			t.Error("expected waitBusy to return false (timeout)")
		}
		if calls != ataTimeout {
			t.Errorf("expected %d calls, got %d", ataTimeout, calls)
		}
	})
}

func TestWaitDRQ(t *testing.T) {
	t.Run("drq ready immediately", func(t *testing.T) {
		ResetMocks()
		defer ResetMocks()
		calls := 0
		MockInb = func(port uint16) byte {
			calls++
			return 0x08 // DRQ ready
		}
		if !waitDRQ() {
			t.Error("expected waitDRQ to return true")
		}
		if calls != 1 {
			t.Errorf("expected 1 call, got %d", calls)
		}
	})

	t.Run("error flag set", func(t *testing.T) {
		ResetMocks()
		defer ResetMocks()
		calls := 0
		MockInb = func(port uint16) byte {
			calls++
			return 0x01 // ERR
		}
		if waitDRQ() {
			t.Error("expected waitDRQ to return false (error)")
		}
		if calls != 1 {
			t.Errorf("expected 1 call, got %d", calls)
		}
	})

	t.Run("timeout", func(t *testing.T) {
		ResetMocks()
		defer ResetMocks()
		calls := 0
		MockInb = func(port uint16) byte {
			calls++
			return 0x00 // Not ready, no error
		}
		if waitDRQ() {
			t.Error("expected waitDRQ to return false (timeout)")
		}
		if calls != ataTimeout {
			t.Errorf("expected %d calls, got %d", ataTimeout, calls)
		}
	})
}

func TestReadSector(t *testing.T) {
	t.Run("successful read", func(t *testing.T) {
		ResetMocks()
		defer ResetMocks()

		inbCalls := 0
		MockInb = func(port uint16) byte {
			inbCalls++
			if inbCalls == 1 {
				return 0x00 // Not busy
			}
			return 0x08 // DRQ ready
		}

		outbPorts := []uint16{}
		outbVals := []byte{}
		MockOutb = func(port uint16, value byte) {
			outbPorts = append(outbPorts, port)
			outbVals = append(outbVals, value)
		}

		inswCalled := false
		MockInsw = func(port uint16, addr *byte, count int) {
			if port != Data {
				t.Errorf("expected insw on Data port, got %x", port)
			}
			if count != 256 {
				t.Errorf("expected 256 words, got %d", count)
			}
			inswCalled = true
		}

		var buf [512]byte
		ok := ReadSector(0x1234567, &buf)
		if !ok {
			t.Error("expected ReadSector to succeed")
		}
		if !inswCalled {
			t.Error("expected insw to be called")
		}

		expectedPorts := []uint16{DriveHead, SecCount, LBALo, LBAMid, LBAHi, StatusCmd}
		if len(outbPorts) != len(expectedPorts) {
			t.Fatalf("expected %d outb calls, got %d", len(expectedPorts), len(outbPorts))
		}
		for i, p := range expectedPorts {
			if outbPorts[i] != p {
				t.Errorf("expected port %x at index %d, got %x", p, i, outbPorts[i])
			}
		}

		// 0x1234567 -> LBALo: 0x67, LBAMid: 0x45, LBAHi: 0x23, DriveHead: 0xE1
		expectedVals := []byte{0xE1, 1, 0x67, 0x45, 0x23, CmdRead}
		for i, v := range expectedVals {
			if outbVals[i] != v {
				t.Errorf("expected value %x at index %d, got %x", v, i, outbVals[i])
			}
		}
	})

	t.Run("busy timeout", func(t *testing.T) {
		ResetMocks()
		defer ResetMocks()

		MockInb = func(port uint16) byte {
			return 0x80 // Busy
		}

		var buf [512]byte
		ok := ReadSector(0x0, &buf)
		if ok {
			t.Error("expected ReadSector to fail on busy timeout")
		}
	})

	t.Run("drq error", func(t *testing.T) {
		ResetMocks()
		defer ResetMocks()

		inbCalls := 0
		MockInb = func(port uint16) byte {
			inbCalls++
			if inbCalls == 1 {
				return 0x00
			}
			return 0x01
		}
		inswCalled := false
		MockInsw = func(port uint16, addr *byte, count int) {
			inswCalled = true
		}

		var buf [512]byte
		if ReadSector(0, &buf) {
			t.Error("expected ReadSector to fail on DRQ error")
		}
		if inswCalled {
			t.Error("expected insw not to be called after DRQ error")
		}
	})
}

func TestWriteSector(t *testing.T) {
	t.Run("successful write", func(t *testing.T) {
		ResetMocks()
		defer ResetMocks()

		inbCalls := 0
		MockInb = func(port uint16) byte {
			inbCalls++
			if inbCalls == 1 {
				return 0x00 // Not busy
			}
			if inbCalls == 2 {
				return 0x08 // DRQ ready
			}
			return 0x00 // Not busy (for flush)
		}

		outbPorts := []uint16{}
		outbVals := []byte{}
		MockOutb = func(port uint16, value byte) {
			outbPorts = append(outbPorts, port)
			outbVals = append(outbVals, value)
		}

		outswCalled := false
		MockOutsw = func(port uint16, addr *byte, count int) {
			if port != Data {
				t.Errorf("expected outsw on Data port, got %x", port)
			}
			if count != 256 {
				t.Errorf("expected 256 words, got %d", count)
			}
			outswCalled = true
		}

		var buf [512]byte
		ok := WriteSector(0x1234567, &buf)
		if !ok {
			t.Error("expected WriteSector to succeed")
		}
		if !outswCalled {
			t.Error("expected outsw to be called")
		}

		// Write sequence: DriveHead, SecCount, LBALo, LBAMid, LBAHi, StatusCmd(CmdWrite)
		// Then outsw
		// Then Flush Cache: StatusCmd(CmdFlush)
		expectedPorts := []uint16{DriveHead, SecCount, LBALo, LBAMid, LBAHi, StatusCmd, StatusCmd}
		if len(outbPorts) != len(expectedPorts) {
			t.Fatalf("expected %d outb calls, got %d", len(expectedPorts), len(outbPorts))
		}
		for i, p := range expectedPorts {
			if outbPorts[i] != p {
				t.Errorf("expected port %x at index %d, got %x", p, i, outbPorts[i])
			}
		}

		expectedVals := []byte{0xE1, 1, 0x67, 0x45, 0x23, CmdWrite, CmdFlush}
		for i, v := range expectedVals {
			if outbVals[i] != v {
				t.Errorf("expected value %x at index %d, got %x", v, i, outbVals[i])
			}
		}
	})

	t.Run("drq timeout", func(t *testing.T) {
		ResetMocks()
		defer ResetMocks()

		inbCalls := 0
		MockInb = func(port uint16) byte {
			inbCalls++
			if inbCalls == 1 {
				return 0x00 // Not busy
			}
			return 0x00 // Not ready (DRQ timeout)
		}

		MockOutb = func(port uint16, value byte) {}

		var buf [512]byte
		ok := WriteSector(0x0, &buf)
		if ok {
			t.Error("expected WriteSector to fail on DRQ timeout")
		}
	})

	t.Run("busy timeout", func(t *testing.T) {
		ResetMocks()
		defer ResetMocks()

		MockInb = func(port uint16) byte { return 0x80 }
		outswCalled := false
		MockOutsw = func(port uint16, addr *byte, count int) {
			outswCalled = true
		}

		var buf [512]byte
		if WriteSector(0, &buf) {
			t.Error("expected WriteSector to fail on busy timeout")
		}
		if outswCalled {
			t.Error("expected outsw not to be called after busy timeout")
		}
	})

	t.Run("flush busy timeout", func(t *testing.T) {
		ResetMocks()
		defer ResetMocks()

		inbCalls := 0
		MockInb = func(port uint16) byte {
			inbCalls++
			if inbCalls == 1 {
				return 0x00
			}
			if inbCalls == 2 {
				return 0x08
			}
			return 0x80
		}
		MockOutsw = func(port uint16, addr *byte, count int) {}

		var buf [512]byte
		if WriteSector(0, &buf) {
			t.Error("expected WriteSector to fail when flush remains busy")
		}
	})
}
