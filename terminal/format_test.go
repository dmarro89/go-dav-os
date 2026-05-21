package terminal

import (
	"math"
	"testing"
)

func formatIntString(v int) string {
	buf, start := FormatInt(v)
	return string(buf[start:])
}

// FormatInt drives PrintInt's decimal output. The cases below cover the
// branches PrintInt used to handle inline (zero, positive, negative) plus
// the two int-range edges that exercised the prior `v = -v` overflow.
func TestFormatInt(t *testing.T) {
	tests := []struct {
		name string
		in   int
		want string
	}{
		{"zero", 0, "0"},
		{"single digit positive", 7, "7"},
		{"single digit negative", -3, "-3"},
		{"two digit positive", 42, "42"},
		{"three digit positive", 123, "123"},
		{"three digit negative", -123, "-123"},
		{"large positive", 9876543210, "9876543210"},
		{"max int64", math.MaxInt64, "9223372036854775807"},
		{"min int64", math.MinInt64, "-9223372036854775808"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatIntString(tt.in)
			if got != tt.want {
				t.Fatalf("FormatInt(%d) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

// FormatInt's output is consumed by PrintInt one byte at a time. Callers
// rely on every byte being a printable ASCII digit (or the leading '-').
// This is implicitly required by `PutRune(rune(buf[i]))` in PrintInt.
func TestFormatIntReturnsAsciiDigitsOnly(t *testing.T) {
	for _, v := range []int{0, 1, -1, 42, -42, 1_000_000, -1_000_000, math.MaxInt64, math.MinInt64} {
		got := formatIntString(v)
		for i, b := range got {
			if i == 0 && b == '-' && v < 0 {
				continue
			}
			if b < '0' || b > '9' {
				t.Fatalf("FormatInt(%d)[%d] = %q, want ASCII digit", v, i, b)
			}
		}
	}
}
