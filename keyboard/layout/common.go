package layout

func lookupKey(keys *[128]rune, sc byte) (rune, bool) {
	if int(sc) >= len(*keys) {
		return 0, false
	}
	r := keys[sc]
	if r == 0 {
		return 0, false
	}
	return r, true
}
