package layout

import "testing"

// TestUSLayout_LetterScancodes verifies the canonical letter scancodes
// produce the expected lowercase letters.
func TestUSLayout_LetterScancodes(t *testing.T) {
	l, _ := GetUS()
	cases := []struct {
		sc   byte
		want rune
	}{
		{0x10, 'q'},
		{0x11, 'w'},
		{0x12, 'e'},
		{0x13, 'r'},
		{0x14, 't'},
		{0x15, 'y'},
		{0x1E, 'a'},
		{0x1F, 's'},
		{0x20, 'd'},
		{0x2C, 'z'},
		{0x32, 'm'},
	}
	for _, tc := range cases {
		got, ok := l.GetKey(tc.sc)
		if !ok {
			t.Errorf("US.GetKey(0x%02X) -> ok=false, want true", tc.sc)
			continue
		}
		if got != tc.want {
			t.Errorf("US.GetKey(0x%02X) = %q, want %q", tc.sc, got, tc.want)
		}
	}
}

// TestITLayout_LetterScancodes verifies the IT layout maps the same letter
// scancodes as the US layout (the visible difference is in shifted symbols,
// not in the unshifted letter positions).
func TestITLayout_LetterScancodes(t *testing.T) {
	l, _ := GetIT()
	cases := []struct {
		sc   byte
		want rune
	}{
		{0x10, 'q'},
		{0x11, 'w'},
		{0x1E, 'a'},
		{0x32, 'm'},
	}
	for _, tc := range cases {
		got, ok := l.GetKey(tc.sc)
		if !ok {
			t.Errorf("IT.GetKey(0x%02X) -> ok=false, want true", tc.sc)
			continue
		}
		if got != tc.want {
			t.Errorf("IT.GetKey(0x%02X) = %q, want %q", tc.sc, got, tc.want)
		}
	}
}

// TestLayouts_DigitScancodes verifies the digit row produces 1..9, 0 on
// both layouts. The shifted symbols differ between US and IT but the
// unshifted digits are the same.
func TestLayouts_DigitScancodes(t *testing.T) {
	cases := []struct {
		sc   byte
		want rune
	}{
		{0x02, '1'},
		{0x03, '2'},
		{0x04, '3'},
		{0x05, '4'},
		{0x06, '5'},
		{0x07, '6'},
		{0x08, '7'},
		{0x09, '8'},
		{0x0A, '9'},
		{0x0B, '0'},
	}
	for name, getter := range map[string]func() (interface{ GetKey(byte) (rune, bool) }, string){
		"US": func() (interface{ GetKey(byte) (rune, bool) }, string) {
			l, n := GetUS()
			return l, n
		},
		"IT": func() (interface{ GetKey(byte) (rune, bool) }, string) {
			l, n := GetIT()
			return l, n
		},
	} {
		l, _ := getter()
		for _, tc := range cases {
			got, ok := l.GetKey(tc.sc)
			if !ok {
				t.Errorf("%s.GetKey(0x%02X) -> ok=false, want true", name, tc.sc)
				continue
			}
			if got != tc.want {
				t.Errorf("%s.GetKey(0x%02X) = %q, want %q", name, tc.sc, got, tc.want)
			}
		}
	}
}

// TestUSLayout_ShiftDigitSymbols verifies the US shifted-digit symbols.
func TestUSLayout_ShiftDigitSymbols(t *testing.T) {
	l, _ := GetUS()
	us, ok := l.(*USLayout)
	if !ok {
		t.Fatalf("expected GetUS to return *USLayout, got %T", l)
	}
	cases := []struct {
		in   rune
		want rune
	}{
		{'1', '!'},
		{'2', '@'},
		{'3', '#'},
		{'4', '$'},
		{'5', '%'},
		{'6', '^'},
		{'7', '&'},
		{'8', '*'},
		{'9', '('},
		{'0', ')'},
	}
	for _, tc := range cases {
		got, ok := us.GetShiftDigitSymbol(tc.in)
		if !ok {
			t.Errorf("US.GetShiftDigitSymbol(%q) -> ok=false, want true", tc.in)
			continue
		}
		if got != tc.want {
			t.Errorf("US.GetShiftDigitSymbol(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// TestITLayout_ShiftDigitSymbols verifies the IT shifted-digit symbols, which
// differ from US (e.g. shift+2 is '"' on IT vs '@' on US, shift+3 is '£' on
// IT vs '#' on US).
func TestITLayout_ShiftDigitSymbols(t *testing.T) {
	l, _ := GetIT()
	it, ok := l.(*ITLayout)
	if !ok {
		t.Fatalf("expected GetIT to return *ITLayout, got %T", l)
	}
	cases := []struct {
		in   rune
		want rune
	}{
		{'1', '!'},
		{'2', '"'},
		{'3', '£'},
		{'4', '$'},
		{'5', '%'},
		{'6', '&'},
		{'7', '/'},
		{'8', '('},
		{'9', ')'},
		{'0', '='},
	}
	for _, tc := range cases {
		got, ok := it.GetShiftDigitSymbol(tc.in)
		if !ok {
			t.Errorf("IT.GetShiftDigitSymbol(%q) -> ok=false, want true", tc.in)
			continue
		}
		if got != tc.want {
			t.Errorf("IT.GetShiftDigitSymbol(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// TestShiftDigitSymbols_DiffBetweenLayouts asserts that at least one shifted
// digit produces a different symbol on IT vs US, so the two layout tables
// can't silently regress to identical mappings.
func TestShiftDigitSymbols_DiffBetweenLayouts(t *testing.T) {
	usLayout, _ := GetUS()
	itLayout, _ := GetIT()
	us := usLayout.(*USLayout)
	it := itLayout.(*ITLayout)

	differs := false
	for _, d := range []rune{'1', '2', '3', '4', '5', '6', '7', '8', '9', '0'} {
		usSym, _ := us.GetShiftDigitSymbol(d)
		itSym, _ := it.GetShiftDigitSymbol(d)
		if usSym != itSym {
			differs = true
			break
		}
	}
	if !differs {
		t.Fatalf("expected at least one shifted-digit symbol to differ between US and IT layouts")
	}
}

// TestLayouts_UnknownScancodesReturnFalse covers the early-return paths in
// lookupKey: scancodes outside the 128-entry table and scancodes inside the
// table whose entry is the zero rune (i.e. unmapped).
func TestLayouts_UnknownScancodesReturnFalse(t *testing.T) {
	for name, l := range map[string]interface {
		GetKey(byte) (rune, bool)
	}{
		"US": func() interface {
			GetKey(byte) (rune, bool)
		} {
			l, _ := GetUS()
			return l
		}(),
		"IT": func() interface {
			GetKey(byte) (rune, bool)
		} {
			l, _ := GetIT()
			return l
		}(),
	} {
		// 0x00 is not mapped in either layout (table entry is the zero rune).
		if r, ok := l.GetKey(0x00); ok {
			t.Errorf("%s.GetKey(0x00) returned (%q, true); want (_, false) for unmapped entry", name, r)
		}
		// 0x7F is not in any layout's filled scancodes.
		if r, ok := l.GetKey(0x7F); ok {
			t.Errorf("%s.GetKey(0x7F) returned (%q, true); want (_, false) for unmapped entry", name, r)
		}
	}
}
