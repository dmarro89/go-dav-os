package terminal

// FormatInt formats v as base-10 ASCII bytes, with a leading '-' for negatives.
// It returns the populated buffer and the offset at which valid bytes begin.
func FormatInt(v int) (buf [20]byte, start int) {
	start = len(buf)
	if v == 0 {
		start--
		buf[start] = '0'
		return buf, start
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

	for u > 0 {
		start--
		buf[start] = byte('0' + (u % 10))
		u /= 10
	}

	if negative {
		start--
		buf[start] = '-'
	}
	return buf, start
}
