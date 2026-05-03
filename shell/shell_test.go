package shell

import (
	"testing"
)

// Test helper to set lineBuf for testing
func setLineBuf(content string) {
	lineLen = len(content)
	for i := 0; i < len(content) && i < maxLine; i++ {
		lineBuf[i] = byte(content[i])
	}
}

// TestIsSpace tests the isSpace helper function
func TestIsSpace(t *testing.T) {
	tests := []struct {
		name     string
		b        byte
		expected bool
	}{
		{"space is space", ' ', true},
		{"tab is space", '\t', true},
		{"newline is not space", '\n', false},
		{"letter is not space", 'a', false},
		{"digit is not space", '0', false},
		{"zero byte is not space", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isSpace(tt.b)
			if result != tt.expected {
				t.Errorf("isSpace(%q) = %v, expected %v", tt.b, result, tt.expected)
			}
		})
	}
}

// TestTrimLeft tests the trimLeft helper function
func TestTrimLeft(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		start    int
		end      int
		expected int
	}{
		{"no spaces", "hello", 0, 5, 0},
		{"leading spaces", "   hello", 0, 8, 3},
		{"leading tabs", "\t\thello", 0, 7, 2},
		{"mixed spaces and tabs", " \t hello", 0, 8, 3},
		{"only spaces", "     ", 0, 5, 5},
		{"space at end", "hello ", 0, 6, 0},
		{"start in middle", "hello world", 5, 11, 6},
		{"start equals end", "hello", 5, 5, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setLineBuf(tt.content)
			result := trimLeft(tt.start, tt.end)
			if result != tt.expected {
				t.Errorf("trimLeft(%d, %d) = %d, expected %d", tt.start, tt.end, result, tt.expected)
			}
		})
	}
}

// TestTrimRight tests the trimRight helper function
func TestTrimRight(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		start    int
		end      int
		expected int
	}{
		{"no trailing spaces", "hello", 0, 5, 5},
		{"trailing spaces", "hello   ", 0, 8, 5},
		{"trailing tabs", "hello\t\t", 0, 7, 5},
		{"mixed spaces and tabs at end", "hello \t ", 0, 8, 5},
		{"only spaces", "     ", 0, 5, 0},
		{"space at start", " hello", 0, 6, 6},
		{"start in middle", "hello world   ", 6, 14, 11},
		{"start equals end", "hello", 5, 5, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setLineBuf(tt.content)
			result := trimRight(tt.start, tt.end)
			if result != tt.expected {
				t.Errorf("trimRight(%d, %d) = %d, expected %d", tt.start, tt.end, result, tt.expected)
			}
		})
	}
}

// TestFirstToken tests the firstToken helper function
func TestFirstToken(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		start    int
		end      int
		expStart int
		expEnd   int
	}{
		{"single word", "hello", 0, 5, 0, 5},
		{"word with spaces after", "hello world", 0, 11, 0, 5},
		{"word with space before and after", " hello world", 1, 12, 1, 6},
		{"empty range", "hello", 0, 0, 0, 0},
		{"space only", "   ", 0, 3, 0, 0},
		{"word starting from middle", "hello world test", 6, 16, 6, 11},
		{"word with tab separator", "hello\tworld", 0, 11, 0, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setLineBuf(tt.content)
			s, e := firstToken(tt.start, tt.end)
			if s != tt.expStart || e != tt.expEnd {
				t.Errorf("firstToken(%d, %d) = (%d, %d), expected (%d, %d)", tt.start, tt.end, s, e, tt.expStart, tt.expEnd)
			}
		})
	}
}

// TestMatchLiteral tests the matchLiteral helper function
func TestMatchLiteral(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		start    int
		end      int
		literal  string
		expected bool
	}{
		{"exact match", "hello", 0, 5, "hello", true},
		{"word in sentence", "hello world", 0, 5, "hello", true},
		{"partial word no match", "hello world", 0, 3, "hello", false},
		{"case sensitive", "Hello", 0, 5, "hello", false},
		{"empty literal", "hello", 0, 0, "", true},
		{"single char", "a", 0, 1, "a", true},
		{"single char mismatch", "a", 0, 1, "b", false},
		{"longer word match", "history", 0, 7, "history", true},
		{"longer word mismatch", "histori", 0, 7, "history", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setLineBuf(tt.content)
			result := matchLiteral(tt.start, tt.end, tt.literal)
			if result != tt.expected {
				t.Errorf("matchLiteral(%d, %d, %q) = %v, expected %v", tt.start, tt.end, tt.literal, result, tt.expected)
			}
		})
	}
}

// TestParseHex64 tests the parseHex64 helper function
func TestParseHex64(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		start    int
		end      int
		expected uint64
		ok       bool
	}{
		{"hex with 0x prefix", "0x1A", 0, 4, 26, true},
		{"hex with 0X prefix", "0X1A", 0, 4, 26, true},
		{"hex without prefix", "1A", 0, 2, 26, true},
		{"lowercase hex", "0xbeef", 0, 6, 0xBEEF, true},
		{"uppercase hex", "0xBEEF", 0, 6, 0xBEEF, true},
		{"mixed case hex", "0xAbCd", 0, 6, 0xABCD, true},
		{"zero", "0x0", 0, 3, 0, true},
		{"large hex", "0xFFFFFFFFFFFFFFFF", 0, 18, 0xFFFFFFFFFFFFFFFF, true},
		{"single digit", "0xF", 0, 3, 15, true},
		{"decimal digits only", "0x123", 0, 5, 0x123, true},
		{"empty input", "", 0, 0, 0, false},
		{"invalid hex char", "0xZZ", 0, 4, 0, false},
		{"only 0x prefix", "0x", 0, 2, 0, false},
		{"hex at offset", "hello0x1A", 5, 9, 26, true},
		{"hex followed by non-hex requires exact boundary", "0x1Aend", 0, 4, 26, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setLineBuf(tt.content)
			result, ok := parseHex64(tt.start, tt.end)
			if ok != tt.ok {
				t.Errorf("parseHex64(%d, %d) ok = %v, expected %v", tt.start, tt.end, ok, tt.ok)
			}
			if ok && result != tt.expected {
				t.Errorf("parseHex64(%d, %d) = %d, expected %d", tt.start, tt.end, result, tt.expected)
			}
		})
	}
}

// TestParseDec tests the parseDec helper function
func TestParseDec(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		start    int
		end      int
		expected int
		ok       bool
	}{
		{"single digit", "5", 0, 1, 5, true},
		{"multiple digits", "123", 0, 3, 123, true},
		{"zero", "0", 0, 1, 0, true},
		{"large number", "999999", 0, 6, 999999, true},
		{"leading non-digit", "a123", 0, 4, 0, false},
		{"trailing non-digit", "123a", 0, 4, 0, false},
		{"space in number", "12 34", 0, 5, 0, false},
		{"empty input", "", 0, 0, 0, false},
		{"number at offset", "hello64", 5, 7, 64, true},
		{"decimal with leading zeros", "00123", 0, 5, 123, true},
		{"negative sign not supported", "-5", 0, 2, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setLineBuf(tt.content)
			result, ok := parseDec(tt.start, tt.end)
			if ok != tt.ok {
				t.Errorf("parseDec(%d, %d) ok = %v, expected %v", tt.start, tt.end, ok, tt.ok)
			}
			if ok && result != tt.expected {
				t.Errorf("parseDec(%d, %d) = %d, expected %d", tt.start, tt.end, result, tt.expected)
			}
		})
	}
}

// TestNextArg tests the nextArg helper function
func TestNextArg(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		start    int
		end      int
		expStart int
		expEnd   int
		expOk    bool
	}{
		{"single arg after space", "hello world", 5, 11, 6, 11, true},
		{"arg with leading spaces", "  hello", 0, 7, 2, 7, true},
		{"multiple spaces between", "hello    world", 5, 14, 9, 14, true},
		{"no more args", "hello ", 5, 6, 0, 0, false},
		{"empty input", "", 0, 0, 0, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setLineBuf(tt.content)
			s, e, ok := nextArg(tt.start, tt.end)
			if ok != tt.expOk || (ok && (s != tt.expStart || e != tt.expEnd)) {
				t.Errorf("nextArg(%d, %d) = (%d, %d, %v), expected (%d, %d, %v)", tt.start, tt.end, s, e, ok, tt.expStart, tt.expEnd, tt.expOk)
			}
		})
	}
}

// TestCopyNameFromRange tests the copyNameFromRange helper function
func TestCopyNameFromRange(t *testing.T) {
	tests := []struct {
		name    string
		content string
		start   int
		end     int
		expLen  int
		expOk   bool
	}{
		{"valid name", "hello", 0, 5, 5, true},
		{"short name", "hi", 0, 2, 2, true},
		{"max length name", "abcdefghijklmnop", 0, 16, 16, true},
		{"too long name", "abcdefghijklmnopq", 0, 17, 0, false},
		{"empty name", "", 0, 0, 0, false},
		{"name at offset", "hello world", 6, 11, 5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setLineBuf(tt.content)
			result, ok := copyNameFromRange(tt.start, tt.end)
			if ok != tt.expOk {
				t.Errorf("copyNameFromRange(%d, %d) ok = %v, expected %v", tt.start, tt.end, ok, tt.expOk)
			}
			if ok && result != tt.expLen {
				t.Errorf("copyNameFromRange(%d, %d) = %d, expected %d", tt.start, tt.end, result, tt.expLen)
			}
		})
	}
}

// TestMinInt tests the minInt helper function
func TestMinInt(t *testing.T) {
	tests := []struct {
		name     string
		a        int
		b        int
		expected int
	}{
		{"a smaller", 5, 10, 5},
		{"b smaller", 10, 5, 5},
		{"equal values", 7, 7, 7},
		{"zero and positive", 0, 5, 0},
		{"negative and positive", -5, 5, -5},
		{"both negative", -5, -10, -10},
		{"large numbers", 1000000, 999999, 999999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := minInt(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("minInt(%d, %d) = %d, expected %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

// TestCopyDataFromRange tests the copyDataFromRange helper function
func TestCopyDataFromRange(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		start    int
		end      int
		expected uint32
	}{
		{"single byte", "a", 0, 1, 1},
		{"multiple bytes", "hello", 0, 5, 5},
		{"data at offset", "hello world", 6, 11, 5},
		{"data within line limit", "hello world test data here", 0, 26, 26},
		{"data at max line boundary", "x", 0, 128, 128},
		{"invalid range", "hello", 5, 2, 0},
		{"empty range", "hello", 0, 0, 0},
		{"end equals start", "hello", 2, 2, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setLineBuf(tt.content)
			result := copyDataFromRange(tt.start, tt.end)
			if result != tt.expected {
				t.Errorf("copyDataFromRange(%d, %d) = %d, expected %d", tt.start, tt.end, result, tt.expected)
			}
		})
	}
}

// TestCalculateDistance tests the calculateDistance (Levenshtein distance) function
func TestCalculateDistance(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		start    int
		end      int
		cmd      string
		expected int
	}{
		{"exact match", "help", 0, 4, "help", 0},
		{"one char difference", "herp", 0, 4, "help", 1},
		{"one char deletion", "hel", 0, 3, "help", 1},
		{"one char insertion", "helpp", 0, 5, "help", 1},
		{"complete mismatch", "abc", 0, 3, "xyz", 3},
		{"empty string", "", 0, 0, "help", 4},
		{"one letter match", "h", 0, 1, "help", 3},
		{"case sensitive", "Help", 0, 4, "help", 1},
		{"transposition chars", "ehpl", 0, 4, "help", 3},
		{"similar words", "history", 0, 7, "history", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setLineBuf(tt.content)
			result := calculateDistance(tt.start, tt.end, tt.cmd)
			if result != tt.expected {
				t.Errorf("calculateDistance(%d, %d, %q) = %d, expected %d", tt.start, tt.end, tt.cmd, result, tt.expected)
			}
		})
	}
}

// TestIntegration_ParseHexAndDec tests integration of hex and decimal parsing
func TestIntegration_ParseHexAndDec(t *testing.T) {
	t.Run("hex address followed by decimal length", func(t *testing.T) {
		setLineBuf("0xB8000 160")
		addrStart, addrEnd := firstToken(0, 11)
		addr, addrOk := parseHex64(addrStart, addrEnd)

		if !addrOk || addr != 0xB8000 {
			t.Errorf("address parsing failed: got %x", addr)
		}

		lenStart, lenEnd, lenOk := nextArg(addrEnd, 11)
		if !lenOk {
			t.Errorf("nextArg failed to find length argument")
		}

		length, lengthOk := parseDec(lenStart, lenEnd)
		if !lengthOk || length != 160 {
			t.Errorf("length parsing failed: got %d", length)
		}
	})
}

// TestIntegration_CommandParsing tests integration of command parsing functions
func TestIntegration_CommandParsing(t *testing.T) {
	t.Run("parse 'echo hello world'", func(t *testing.T) {
		setLineBuf("  echo hello world  ")
		start := trimLeft(0, 20)
		end := trimRight(start, 20)

		cmdStart, cmdEnd := firstToken(start, end)
		if !matchLiteral(cmdStart, cmdEnd, "echo") {
			t.Errorf("failed to match 'echo' command")
		}

		msgStart, msgEnd, ok := nextArg(cmdEnd, end)
		if !ok || msgStart != 7 || msgEnd != 12 {
			t.Errorf("failed to parse message argument: got (%d, %d), expected (7, 12)", msgStart, msgEnd)
		}
	})
}

// TestEdgeCases tests edge cases and boundary conditions
func TestEdgeCases(t *testing.T) {
	t.Run("trimLeft and trimRight with only spaces", func(t *testing.T) {
		setLineBuf("     ")
		start := trimLeft(0, 5)
		end := trimRight(start, 5)
		if start != 5 || end != 5 {
			t.Errorf("expected both to be 5, got start=%d, end=%d", start, end)
		}
	})

	t.Run("parseHex64 with all F's", func(t *testing.T) {
		setLineBuf("0xffffffffffffffff")
		val, ok := parseHex64(0, 18)
		if !ok || val != 0xffffffffffffffff {
			t.Errorf("parseHex64 failed for all F's")
		}
	})

	t.Run("parseDec overflow behavior", func(t *testing.T) {
		setLineBuf("999999999999999999")
		_, ok := parseDec(0, 18)
		// Should parse successfully but may overflow - just verify it doesn't crash
		if !ok {
			t.Errorf("parseDec should handle large numbers")
		}
	})
}
