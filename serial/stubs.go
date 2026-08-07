//go:build !gccgo && !testing

package serial

func initPort() bool               { return false }
func tryWriteByte(value byte) bool { return false }
func tryReadByte() (byte, bool)    { return 0, false }
