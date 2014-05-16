package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"strings"
	"unicode/utf16"

	"xy/text"
)

var le = binary.LittleEndian

func main() {
	in := os.Stdin
	if len(os.Args) > 1 {
		f, err := os.Open(os.Args[1])
		if err != nil {
			fmt.Println(err)
			return
		}
		defer f.Close()
		in = f
	}
	ss, err := text.Read(in)
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, s := range ss {
		fmt.Printf("%q\n", s)
	}
}

func do(f *os.File) error {
	var err error
	var header struct {
		Sections uint16
		Lines    uint16
		Size     uint32
	}
	err = binary.Read(f, le, &header)
	if err != nil {
		return err
	}

	if header.Sections != 1 {
		return errors.New("too many sections!")
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
	err = binary.Read(f, le, junk)
	if err != nil {
		return err
	}
	err = binary.Read(f, le, entries)
	if err != nil {
		return err
	}
	err = binary.Read(f, le, chars)
	if err != nil {
		return err
	}
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
		fmt.Printf("%q\n", s)
	}
	return nil
}
