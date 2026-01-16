package terminal

import "unsafe"

func outb(port uint16, value byte)
func debugChar(c byte)

const (
	VGAWidth  = 80
	VGAHeight = 25

	vgaCursorIndexPort uint16 = 0x3D4
	vgaCursorDataPort  uint16 = 0x3D5
)

const videoMemoryAddr uintptr = 0xB8000

func getVidMem() *[VGAHeight][VGAWidth][2]byte {
	return (*[VGAHeight][VGAWidth][2]byte)(unsafe.Pointer(videoMemoryAddr))
}

const (
	ColorBlack     = 0
	ColorLightGrey = 7
)

var (
	column int
	row    int
	color  byte
	vidMem *[VGAHeight][VGAWidth][2]byte
)

func Init() {
	vidMem = getVidMem()
	color = makeColor(ColorLightGrey, ColorBlack)
	column = 0
	row = 0
	Clear()
}

func makeColor(fg, bg byte) byte {
	return fg | (bg << 4)
}

func Clear() {
	for r := 0; r < VGAHeight; r++ {
		for c := 0; c < VGAWidth; c++ {
			vidMem[r][c][0] = ' '
			vidMem[r][c][1] = color
		}
	}
	column = 0
	row = 0
	updateCursor()
}

func PutRune(ch rune) {
	putRune(ch)
}

func putRune(ch rune) {
	if ch == '\b' {
		Backspace()
		return
	}

	if ch == '\n' {
		column = 0
		row++
		if row >= VGAHeight {
			scroll()
			row = VGAHeight - 1
		}
		updateCursor()
		return
	}

	vidMem[row][column][0] = byte(ch)
	vidMem[row][column][1] = color

	debugChar(byte(ch))

	column++
	if column >= VGAWidth {
		column = 0
		row++
		if row >= VGAHeight {
			scroll()
			row = VGAHeight - 1
		}
	}
	updateCursor()
}

func scroll() {
	for r := 1; r < VGAHeight; r++ {
		for c := 0; c < VGAWidth; c++ {
			vidMem[r-1][c] = vidMem[r][c]
		}
	}

	last := VGAHeight - 1
	for c := 0; c < VGAWidth; c++ {
		vidMem[last][c][0] = ' '
		vidMem[last][c][1] = color
	}
}

func Print(s string) {
	for i := 0; i < len(s); i++ {
		putRune(rune(s[i]))
	}
}

func PrintAt(col, row int, s string) {
	for i := 0; i < len(s); i++ {
		putRuneAt(col+i, row, rune(s[i]))
	}
}

func putRuneAt(col, currRow int, ch rune) {
	if col < 0 || col >= VGAWidth {
		return
	}
	if currRow < 0 || currRow >= VGAHeight {
		return
	}

	if ch == '\n' {
		column = 0
		row = currRow + 1
		if row >= VGAHeight {
			scroll()
			row = VGAHeight - 1
		}
		updateCursor()
		return
	}

	vidMem[currRow][col][0] = byte(ch)
	vidMem[currRow][col][1] = color

	column = col + 1
	row = currRow
	if column >= VGAWidth {
		column = 0
		row = currRow + 1
		if row >= VGAHeight {
			scroll()
			row = VGAHeight - 1
		}
	}

	updateCursor()
}

func Backspace() {
	if column > 0 {
		column--
		vidMem[row][column][0] = ' '
		vidMem[row][column][1] = color
	} else {
		if row > 0 {
			row--
			column = VGAWidth - 1
			vidMem[row][column][0] = ' '
			vidMem[row][column][1] = color
		}
	}
	updateCursor()
}

func updateCursor() {
	pos := uint16(row*VGAWidth + column)

	outb(vgaCursorIndexPort, 0x0F)
	outb(vgaCursorDataPort, byte(pos&0xFF))

	outb(vgaCursorIndexPort, 0x0E)
	outb(vgaCursorDataPort, byte((pos>>8)&0xFF))
}
