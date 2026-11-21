package main

import "unicode/utf8"

// A small byte reader for Unix terminal escape sequences. Bracketed paste is
// consumed as data, not as keypresses (so pasted newlines never run commands).
type keyReader struct {
	nextByte func() (byte, bool, error)
	pasting  bool
	endMatch int
}

const pasteEnd = "\x1b[201~"

func (k *keyReader) next() (string, rune, error) {
	b, ok, err := k.nextByte()
	if err != nil || !ok {
		return "", 0, err
	}
	if k.pasting {
		if b == pasteEnd[k.endMatch] {
			k.endMatch++
			if k.endMatch == len(pasteEnd) {
				k.pasting, k.endMatch = false, 0
			}
		} else if b == pasteEnd[0] {
			k.endMatch = 1
		} else {
			k.endMatch = 0
		}
		return "", 0, nil
	}
	switch b {
	case 0x00:
		return "ctrl+space", 0, nil
	case 0x03:
		return "ctrl+c", 0, nil
	case 0x11:
		return "ctrl+q", 0, nil
	case 0x15:
		return "ctrl+u", 0, nil
	case 0x10:
		return "tab-left", 0, nil // Ctrl+P: move active tab left
	case 0x0e:
		return "tab-right", 0, nil // Ctrl+N: move active tab right
	case '?':
		return "help", 0, nil
	case '\r', '\n':
		return "enter", 0, nil
	case 0x08, 0x7f:
		return "backspace", 0, nil
	case 0x1b:
		second, ok, err := k.nextByte()
		if err != nil || !ok {
			return "escape", 0, err
		}
		if second != '[' && second != 'O' {
			return "escape", 0, nil
		}
		// A CSI sequence ends with a final byte in the ASCII range 0x40–0x7E.
		// Cap the bytes, so malformed sequences cannot block indefinitely.
		var seq []byte
		for len(seq) < 12 {
			c, ok, err := k.nextByte()
			if err != nil || !ok {
				return "", 0, err
			}
			seq = append(seq, c)
			if c >= 0x40 && c <= 0x7e {
				break
			}
		}
		if second == '[' {
			switch string(seq) {
			case "A":
				return "up", 0, nil
			case "B":
				return "down", 0, nil
			case "C":
				return "right", 0, nil
			case "D":
				return "left", 0, nil
			case "1;5D", "1;3D":
				return "tab-left", 0, nil // optional Ctrl/Alt+Left
			case "1;5C", "1;3C":
				return "tab-right", 0, nil // optional Ctrl/Alt+Right
			case "200~":
				k.pasting, k.endMatch = true, 0
			}
		} else {
			switch string(seq) {
			case "A":
				return "up", 0, nil
			case "B":
				return "down", 0, nil
			case "C":
				return "right", 0, nil
			case "D":
				return "left", 0, nil
			}
		}
		return "", 0, nil
	}
	if b < 0x20 {
		return "", 0, nil
	}
	if b < utf8.RuneSelf {
		return "rune", rune(b), nil
	}
	buf := []byte{b}
	for !utf8.FullRune(buf) && len(buf) < utf8.UTFMax {
		c, ok, err := k.nextByte()
		if err != nil || !ok {
			return "", 0, err
		}
		buf = append(buf, c)
	}
	r, size := utf8.DecodeRune(buf)
	if r == utf8.RuneError && size == 1 {
		return "", 0, nil
	}
	return "rune", r, nil
}
