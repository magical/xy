package text

import (
	"encoding/binary"
	"errors"
	"io"
	"strconv"
	"unicode/utf8"
	"unicode/utf16"
)

var le = binary.LittleEndian

func ReadRaw(r io.Reader) ([][]uint16, error) {
	var err error
	var header struct {
		Sections uint16
		Lines    uint16
		Size     uint32
	}
	err = binary.Read(r, le, &header)
	if err != nil {
		return nil, err
	}

	if header.Sections != 1 {
		return nil, errors.New("text: too many sections!")
	}

	//fmt.Println(header)
	type entry struct {
		Offset uint32
		Length uint16
		_ uint16 // unknown
	}
	var junk = make([]uint32, header.Sections+2)
	var entries = make([]entry, header.Lines)
	textOff := binary.Size(entries) + 4
	remaining := int(header.Size) - binary.Size(entries) - 4
	//fmt.Printf("%x %x\n", textStart, remaining)
	var chars = make([]uint16, remaining/2)
	err = binary.Read(r, le, junk)
	if err != nil {
		return nil, err
	}
	err = binary.Read(r, le, entries)
	if err != nil {
		return nil, err
	}
	err = binary.Read(r, le, chars)
	if err != nil {
		return nil, err
	}
	var ss [][]uint16
	for _, e := range entries {
		off := int(e.Offset) - textOff
		//fmt.Println(e, off/2, len(chars))
		if off/2 + int(e.Length) > len(chars) {
			return nil, errors.New("text: offset out of range")
		}
		chars := chars[off/2:][:e.Length]
		//fmt.Printf("%x\n", chars)
		key := chars[len(chars)-1]
		for i := len(chars); i > 0; i-- {
			chars[i-1] ^= key
			key = (key>>3 | key<<13)
		}
		ss = append(ss, chomp(chars))
	}
	return ss, nil
}

func Read(r io.Reader) ([]string, error) {
	textsRaw, err := ReadRaw(r)
	if err != nil {
		return nil, err
	}
	texts := make([]string, 0, len(textsRaw))
	for _, s := range textsRaw {
		texts = append(texts, string(utf16.Decode(s)))
	}
	return texts, nil
}

func chomp(s []uint16) []uint16 {
	if len(s) > 0 && s[len(s)-1] == 0 {
		return s[:len(s)-1]
	}
	return s
}

func trimRight(s []uint16) []uint16 {
	i := len(s)
	for i > 0 && s[i-1] == 0 {
		i--
	}
	return s[:i]
}

const lowerhex = "0123456789abcdef"

func Escape(s []uint16) string {
	return escape(s, false)
}

func escape(s []uint16, ASCIIonly bool) string {
	var runeTmp [utf8.UTFMax]byte
	buf := make([]byte, 0, 3*len(s)/2) // Try to avoid more allocations.
	for width := 0; len(s) > 0; s = s[width:] {
		r := rune(s[0])
		width = 1
		if r == 0x10 && len(s) > 1{
			width = int(s[1]) + 2
			if width > len(s) {
				width = len(s)
			}
			buf = append(buf, `\e`...)
			buf = strconv.AppendInt(buf, int64(s[1]), 10)
			buf = append(buf, '{')
			for i := 2; i < width; i++ {
				if i != 2 {
					buf = append(buf, ',')
				}
				buf = append(buf, lowerhex[s[i]>>12&0xF])
				buf = append(buf, lowerhex[s[i]>>8&0xF])
				buf = append(buf, lowerhex[s[i]>>4&0xF])
				buf = append(buf, lowerhex[s[i]&0xF])
			}
			buf = append(buf, '}')
			continue
		}
		if width == 1 && r == utf8.RuneError {
			buf = append(buf, `\x`...)
			buf = append(buf, lowerhex[s[0]>>4])
			buf = append(buf, lowerhex[s[0]&0xF])
			continue
		}
		if ASCIIonly {
			if r < utf8.RuneSelf && strconv.IsPrint(r) {
				buf = append(buf, byte(r))
				continue
			}
		} else if strconv.IsPrint(r) {
			n := utf8.EncodeRune(runeTmp[:], r)
			buf = append(buf, runeTmp[:n]...)
			continue
		}
		switch r {
		//case '\a':
		//	buf = append(buf, `\a`...)
		//case '\b':
		//	buf = append(buf, `\b`...)
		//case '\f':
		//	buf = append(buf, `\f`...)
		case '\n':
			buf = append(buf, `\n`...)
		case '\r':
			buf = append(buf, `\r`...)
		//case '\t':
		//	buf = append(buf, `\t`...)
		//case '\v':
		//	buf = append(buf, `\v`...)
		default:
			switch {
			case r < ' ':
				buf = append(buf, `\x`...)
				buf = append(buf, lowerhex[s[0]>>4])
				buf = append(buf, lowerhex[s[0]&0xF])
			case r > utf8.MaxRune:
				r = 0xFFFD
				fallthrough
			case r < 0x10000:
				buf = append(buf, `\u`...)
				for s := 12; s >= 0; s -= 4 {
					buf = append(buf, lowerhex[r>>uint(s)&0xF])
				}
			default:
				buf = append(buf, `\U`...)
				for s := 28; s >= 0; s -= 4 {
					buf = append(buf, lowerhex[r>>uint(s)&0xF])
				}
			}
		}
	}
	return string(buf)
}
