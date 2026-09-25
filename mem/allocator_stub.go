//go:build !gccgo

package mem

var testingBitmapBacking []byte

func bitmapBytePtr(off uint64) *byte {
	if off >= uint64(len(testingBitmapBacking)) {
		newSize := off + 1024
		newBacking := make([]byte, newSize)
		copy(newBacking, testingBitmapBacking)
		testingBitmapBacking = newBacking
	}
	return &testingBitmapBacking[off]
}
