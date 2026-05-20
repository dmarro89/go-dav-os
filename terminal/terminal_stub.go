//go:build !gccgo && !testing

package terminal

var output string

func Init() {
	output = ""
}

func Clear() {
	output = ""
}

func PutRune(ch rune) {
	output += string(ch)
}

func Print(s string) {
	output += s
}

func PrintAt(col, row int, s string) {
	Print(s)
}

func Backspace() {
	if len(output) > 0 {
		output = output[:len(output)-1]
	}
}

func PrintInt(v int) {
	if v < 0 {
		PutRune('-')
		v = -v
	}
	if v == 0 {
		PutRune('0')
		return
	}
	var buf [20]byte
	i := 0
	val := uint64(v)
	for val > 0 {
		buf[i] = byte('0' + (val % 10))
		val /= 10
		i++
	}
	for i > 0 {
		i--
		PutRune(rune(buf[i]))
	}
}

func PrintHex(v uint64) {
	Print("0x")
	if v == 0 {
		PutRune('0')
		return
	}
	digits := "0123456789ABCDEF"
	var buf [16]byte
	i := 0
	for v > 0 {
		buf[i] = digits[v%16]
		v /= 16
		i++
	}
	for i > 0 {
		i--
		PutRune(rune(buf[i]))
	}
}

func ResetOutputForTesting() {
	output = ""
}

func OutputForTesting() string {
	return output
}
