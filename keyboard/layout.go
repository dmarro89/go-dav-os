//go:build !testing

package keyboard

type Layout interface {
	GetKey(byte) (rune, bool)
	GetShiftDigitSymbol(r rune) (rune, bool)
}
