package text

import (
	"encoding/binary"
	"errors"
	"io"
	"strings"
	"unicode/utf16"
)

var le = binary.LittleEndian

func Read(r io.Reader) ([]string, error) {
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
	var ss []string
	for _, e := range entries {
		off := int(e.Offset) - textOff
		//fmt.Println(e, off/2, len(chars))
		chars := chars[off/2:][:e.Length]
		//fmt.Printf("%x\n", chars)
		key := chars[len(chars)-1]
		for i := len(chars); i > 0; i-- {
			chars[i-1] ^= key
			key = (key>>3 | key<<13)
		}
		s := string(utf16.Decode(chars))
		s = strings.TrimRight(s, "\x00")
		ss = append(ss, s)
	}
	return ss, nil
}
