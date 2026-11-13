//go:build testing

package ata

var (
	MockInb   func(port uint16) byte
	MockOutb  func(port uint16, value byte)
	MockInsw  func(port uint16, addr *byte, count int)
	MockOutsw func(port uint16, addr *byte, count int)
)

func inb(port uint16) byte {
	if MockInb != nil {
		return MockInb(port)
	}
	return 0
}

func outb(port uint16, value byte) {
	if MockOutb != nil {
		MockOutb(port, value)
	}
}

func insw(port uint16, addr *byte, count int) {
	if MockInsw != nil {
		MockInsw(port, addr, count)
	}
}

func outsw(port uint16, addr *byte, count int) {
	if MockOutsw != nil {
		MockOutsw(port, addr, count)
	}
}

// ResetMocks clears all mock functions.
func ResetMocks() {
	MockInb = nil
	MockOutb = nil
	MockInsw = nil
	MockOutsw = nil
}
