//go:build testing

package terminal

var vgaBuffer *[25][80]uint16
var cursorRow int
var cursorCol int
var output string

func outb(port uint16, value byte) {}

func debugChar(c byte) {}

func Init() {
	vgaBuffer = new([25][80]uint16)
	cursorRow = 0
	cursorCol = 0
	output = ""
}

func Clear() {
	if vgaBuffer != nil {
		for row := 0; row < 25; row++ {
			for col := 0; col < 80; col++ {
				vgaBuffer[row][col] = 0
			}
		}
	}
	cursorRow = 0
	cursorCol = 0
}

func PutRune(ch rune) {
	output += string(ch)
}

func Print(s string) {
	output += s
}

func PrintAt(col, row int, s string) {}

func Backspace() {
	if len(output) > 0 {
		output = output[:len(output)-1]
	}
}

func PrintInt(v int) {}

func ResetOutputForTesting() {
	output = ""
}

func OutputForTesting() string {
	return output
}
