//go:build testing

package keyboard

import "testing"

type testLayout struct{}

func (testLayout) GetKey(sc byte) (rune, bool) {
	switch sc {
	case 0x02:
		return '1', true
	case 0x0C:
		return '-', true
	case 0x1E:
		return 'a', true
	default:
		return 0, false
	}
}

func (testLayout) GetShiftSymbol(r rune) (rune, bool) {
	switch r {
	case '1':
		return '!', true
	case '-':
		return '_', true
	default:
		return 0, false
	}
}

func resetKeyboardState() {
	leftShiftDown = false
	rightShiftDown = false
	capsLockOn = false
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

func TestTranslateScancodeNonLetterModifiers(t *testing.T) {
	tests := []struct {
		name  string
		setup func()
		sc    byte
		want  rune
	}{
		{name: "digit", sc: 0x02, want: '1'},
		{
			name: "caps digit unchanged",
			setup: func() {
				translateScancode(scCapsLockDown)
			},
			sc:   0x02,
			want: '1',
		},
		{name: "hyphen", sc: 0x0C, want: '-'},
		{
			name: "shift hyphen underscore",
			setup: func() {
				translateScancode(scLeftShiftDown)
			},
			sc:   0x0C,
			want: '_',
		},
		{
			name: "caps hyphen unchanged",
			setup: func() {
				translateScancode(scCapsLockDown)
			},
			sc:   0x0C,
			want: '-',
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetKeyboardState()
			if tt.setup != nil {
				tt.setup()
			}

			got, ok := translateScancode(tt.sc)
			if !ok {
				t.Fatalf("translateScancode(0x%02X) -> ok=false, want true", tt.sc)
			}
			if got != tt.want {
				t.Fatalf("translateScancode(0x%02X) = %q, want %q", tt.sc, got, tt.want)
			}
		})
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
