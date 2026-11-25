//go:build testing

package keyboard

import "testing"

type testLayout struct{}

func (testLayout) GetKey(sc byte) (rune, bool) {
	switch sc {
	case 0x02:
		return '1', true
	case 0x1E:
		return 'a', true
	default:
		return 0, false
	}
}

func (testLayout) GetShiftDigitSymbol(r rune) (rune, bool) {
	if r == '1' {
		return '!', true
	}
	return 0, false
}

func resetKeyboardState() {
	leftShiftDown = false
	rightShiftDown = false
	capsLockOn = false
	extendedPrefix = false
	SetLayout(testLayout{})
}

func TestTranslateScancodeLetterModifiers(t *testing.T) {
	tests := []struct {
		name  string
		setup func()
		want  rune
	}{
		{
			name: "lowercase",
			want: 'a',
		},
		{
			name: "shift uppercase",
			setup: func() {
				translateScancode(scLeftShiftDown)
			},
			want: 'A',
		},
		{
			name: "caps uppercase",
			setup: func() {
				translateScancode(scCapsLockDown)
			},
			want: 'A',
		},
		{
			name: "shift caps lowercase",
			setup: func() {
				translateScancode(scLeftShiftDown)
				translateScancode(scCapsLockDown)
			},
			want: 'a',
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetKeyboardState()
			if tt.setup != nil {
				tt.setup()
			}

			got, ok := translateScancode(0x1E)
			if !ok {
				t.Fatalf("translateScancode(0x1E) -> ok=false, want true")
			}
			if got != tt.want {
				t.Fatalf("translateScancode(0x1E) = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTranslateScancodeShiftedDigitUsesLayoutSymbol(t *testing.T) {
	resetKeyboardState()
	translateScancode(scRightShiftDown)

	got, ok := translateScancode(0x02)
	if !ok {
		t.Fatalf("translateScancode(0x02) -> ok=false, want true")
	}
	if got != '!' {
		t.Fatalf("translateScancode(0x02) = %q, want %q", got, '!')
	}
}

func TestTranslateScancodeShiftReleaseClearsModifier(t *testing.T) {
	resetKeyboardState()

	if r, ok := translateScancode(scLeftShiftDown); ok {
		t.Fatalf("translateScancode(scLeftShiftDown) = (%q, true), want (_, false)", r)
	}
	got, ok := translateScancode(0x1E)
	if !ok {
		t.Fatalf("translateScancode(0x1E) with shift down -> ok=false, want true")
	}
	if got != 'A' {
		t.Fatalf("translateScancode(0x1E) with shift down = %q, want %q", got, 'A')
	}

	if r, ok := translateScancode(scLeftShiftUp); ok {
		t.Fatalf("translateScancode(scLeftShiftUp) = (%q, true), want (_, false)", r)
	}
	got, ok = translateScancode(0x1E)
	if !ok {
		t.Fatalf("translateScancode(0x1E) after shift release -> ok=false, want true")
	}
	if got != 'a' {
		t.Fatalf("translateScancode(0x1E) after shift release = %q, want %q", got, 'a')
	}
}

func TestTranslateScancodeIgnoresKeyReleasesAndUnmappedKeys(t *testing.T) {
	tests := []struct {
		name string
		sc   byte
	}{
		{name: "letter release", sc: 0x9E},
		{name: "unmapped scancode", sc: 0x7F},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetKeyboardState()

			if r, ok := translateScancode(tt.sc); ok {
				t.Fatalf("translateScancode(0x%02X) = (%q, true), want (_, false)", tt.sc, r)
			}
		})
	}
}

func TestKeyConstants(t *testing.T) {
	if KeyUp != 0x1000 {
		t.Fatalf("KeyUp = 0x%X, want 0x1000", KeyUp)
	}
	if KeyDown != 0x1001 {
		t.Fatalf("KeyDown = 0x%X, want 0x1001", KeyDown)
	}
}

func TestTranslateScancodeArrowsNeedExtendedPrefix(t *testing.T) {
	tests := []struct {
		name  string
		codes []byte
		want  rune
		ok    bool
	}{
		{"extended up", []byte{0xE0, 0x48}, KeyUp, true},
		{"extended down", []byte{0xE0, 0x50}, KeyDown, true},
		{"bare 0x48 (keypad 8) is not history", []byte{0x48}, 0, false},
		{"bare 0x50 (keypad 2) is not history", []byte{0x50}, 0, false},
		{"extended up release is ignored", []byte{0xE0, 0xC8}, 0, false},
		{"extended down release is ignored", []byte{0xE0, 0xD0}, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetKeyboardState()
			var got rune
			var ok bool
			for _, sc := range tt.codes {
				got, ok = translateScancode(sc)
			}
			if got != tt.want || ok != tt.ok {
				t.Fatalf("translateScancode(% X) = (%q, %v), want (%q, %v)", tt.codes, got, ok, tt.want, tt.ok)
			}
		})
	}
}

func TestTranslateScancodePrefixAppliesToOneScancode(t *testing.T) {
	resetKeyboardState()
	translateScancode(0xE0)
	translateScancode(0x1E) // consumes the prefix
	if got, ok := translateScancode(0x48); ok {
		t.Fatalf("bare 0x48 after a consumed prefix = (%q, true), want ignored", got)
	}
}
