package keyboard

type Layout [128]rune

var LayoutIT = Layout{
	0x02: '1',
	0x03: '2',
	0x04: '3',
	0x05: '4',
	0x06: '5',
	0x07: '6',
	0x08: '7',
	0x09: '8',
	0x0A: '9',
	0x0B: '0',

	0x10: 'q',
	0x11: 'w',
	0x12: 'e',
	0x13: 'r',
	0x14: 't',
	0x15: 'y',
	0x16: 'u',
	0x17: 'i',
	0x18: 'o',
	0x19: 'p',

	0x1E: 'a',
	0x1F: 's',
	0x20: 'd',
	0x21: 'f',
	0x22: 'g',
	0x23: 'h',
	0x24: 'j',
	0x25: 'k',
	0x26: 'l',

	0x2C: 'z',
	0x2D: 'x',
	0x2E: 'c',
	0x2F: 'v',
	0x30: 'b',
	0x31: 'n',
	0x32: 'm',

	0x39: ' ',
	0x1C: '\n',
	0x0E: '\b',
}

func isASCIILetter(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

func isASCIIDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func toUpperASCII(r rune) rune {
	if r >= 'a' && r <= 'z' {
		return r - 'a' + 'A'
	}
	return r
}

func toLowerASCII(r rune) rune {
	if r >= 'A' && r <= 'Z' {
		return r - 'A' + 'a'
	}
	return r
}

func toSymbol(r rune) (rune, bool) {
	switch r {
	case '1':
		return '!', true
	case '2':
		return '"', true
	case '3':
		return '£', true
	case '4':
		return '$', true
	case '5':
		return '%', true
	case '6':
		return '&', true
	case '7':
		return '/', true
	case '8':
		return '(', true
	case '9':
		return ')', true
	case '0':
		return '=', true
	default:
		return 0, false
	}
}
