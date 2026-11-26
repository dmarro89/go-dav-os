//go:build !testing

package keyboard

type Layout interface {
	GetKey(byte) (rune, bool)
	GetShiftSymbol(r rune) (rune, bool)
}
