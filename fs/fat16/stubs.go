//go:build testing

package fat16

var (
	BytesPerSec uint16
	SecPerClust uint8
	ReservedSec uint16
	NumFATs     uint8
	RootEntCnt  uint16
	TotSec16    uint16
	FatSz16     uint16
	initialized bool
)

const DirEntrySize = 32

func Init() bool {
	initialized = true
	return true
}

func Format() bool {
	return true
}

func Info() {}

func ListDir() {}

func CreateFile(name *[8]byte, ext *[3]byte, data *[512]byte, dataLen uint32) bool {
	return false
}

func ReadFile(name *[8]byte, ext *[3]byte, outBuf *[512]byte) (uint32, bool) {
	return 0, false
}
