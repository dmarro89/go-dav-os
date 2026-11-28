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

const (
	DirEntrySize = 32

	// Status error codes for FAT16 operations
	StatusOK          = 0
	ErrNotInitialized = 1
	ErrFileNotFound   = 2
	ErrFileExists     = 3
	ErrNoFreeClusters = 4
	ErrDirectoryFull  = 5
	ErrInvalidName    = 6
	ErrDiskIO         = 7
)

func Init() bool {
	initialized = true
	return true
}

func Format() bool {
	return true
}

func Info() {}

func ListDir() {}

var (
	mockCreateFileErr = ErrFileNotFound
	mockReadFileErr   = ErrFileNotFound
	mockReadFileSize  uint32
)

func SetMockCreateFileErr(err int) {
	mockCreateFileErr = err
}

func SetMockReadFile(size uint32, err int) {
	mockReadFileSize = size
	mockReadFileErr = err
}

func SetInitializedForTesting(init bool) {
	initialized = init
}

func ResetForTesting() {
	initialized = true
	mockCreateFileErr = ErrFileNotFound
	mockReadFileErr = ErrFileNotFound
	mockReadFileSize = 0
}

func CreateFileWithErr(name *[8]byte, ext *[3]byte, data *[512]byte, dataLen uint32) int {
	if !initialized {
		return ErrNotInitialized
	}
	if name == nil || name[0] == ' ' || name[0] == 0 {
		return ErrInvalidName
	}
	return mockCreateFileErr
}

func CreateFile(name *[8]byte, ext *[3]byte, data *[512]byte, dataLen uint32) bool {
	return CreateFileWithErr(name, ext, data, dataLen) == StatusOK
}

func ReadFileWithErr(name *[8]byte, ext *[3]byte, outBuf *[512]byte) (uint32, int) {
	if !initialized {
		return 0, ErrNotInitialized
	}
	if name == nil || name[0] == ' ' || name[0] == 0 {
		return 0, ErrInvalidName
	}
	return mockReadFileSize, mockReadFileErr
}

func ReadFile(name *[8]byte, ext *[3]byte, outBuf *[512]byte) (uint32, bool) {
	size, err := ReadFileWithErr(name, ext, outBuf)
	return size, err == StatusOK
}
