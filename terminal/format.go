package terminal

// FormatInt formats v as a base-10 string of bytes, with a leading '-' for
// negative values.
//
// PrintInt and the testing stub both delegate the formatting work to this
// function so the integer-to-decimal conversion can be tested without
// touching VGA memory or the I/O ports backing PutRune.
func FormatInt(v int) []byte {
	if v == 0 {
		return []byte{'0'}
	}

	negative := v < 0
	// Convert to uint64 without overflowing on the most-negative value.
	// `-MinInt` would wrap, so we negate (v+1) and add one to the unsigned
	// magnitude instead.
	var u uint64
	if negative {
		u = uint64(-(v + 1)) + 1
	} else {
		u = uint64(v)
	}

	// 20 bytes covers every uint64 decimal representation; +1 byte for the
	// optional '-' sign.
	var rev [20]byte
	n := 0
	for u > 0 {
		rev[n] = byte('0' + (u % 10))
		u /= 10
		n++
	}

	out := make([]byte, 0, n+1)
	if negative {
		out = append(out, '-')
	}
	for i := n - 1; i >= 0; i-- {
		out = append(out, rev[i])
	}
	return out
}
