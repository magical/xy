package main

import (
	"fmt"
	"os"
	"log"
	"path/filepath"
	"strconv"
	"unicode/utf8"

	"xy/garc"
	"xy/text"
)

func main() {
	filename := os.Args[1]
	outdir := os.Args[2]
	err := do(filename, outdir)
	if err != nil {
		log.Print(err)
	}
}

func do(filename string, outdir string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	gfiles, err := garc.Files(f)
	if err != nil {
		return err
	}
	for gnum, gfile := range gfiles {
		ss, err := text.Read(gfile)
		if err != nil {
			return fmt.Errorf("%s %d: %s", filename, gnum, err)
		}
		outname := fmt.Sprint(gnum)
		out, err := os.Create(filepath.Join(outdir, outname))
		if err != nil {
			return fmt.Errorf("%s %d: %s", filename, gnum, err)
		}
		defer out.Close()
		for _, s := range ss {
			out.WriteString(quoteWith(s, '\x00', false))
			out.WriteString("\n")
		}
	}
	return nil
}

const lowerhex = "0123456789abcdef"

func quoteWith(s string, quote byte, ASCIIonly bool) string {
	var runeTmp [utf8.UTFMax]byte
	buf := make([]byte, 0, 3*len(s)/2) // Try to avoid more allocations.
	//buf = append(buf, quote)
	for width := 0; len(s) > 0; s = s[width:] {
		r := rune(s[0])
		width = 1
		if r >= utf8.RuneSelf {
			r, width = utf8.DecodeRuneInString(s)
		}
		if width == 1 && r == utf8.RuneError {
			buf = append(buf, `\x`...)
			buf = append(buf, lowerhex[s[0]>>4])
			buf = append(buf, lowerhex[s[0]&0xF])
			continue
		}
		/*
		if r == rune(quote) || r == '\\' { // always backslashed
			buf = append(buf, '\\')
			buf = append(buf, byte(r))
			continue
		}
		*/
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
		case '\a':
			buf = append(buf, `\a`...)
		case '\b':
			buf = append(buf, `\b`...)
		case '\f':
			buf = append(buf, `\f`...)
		case '\n':
			buf = append(buf, `\n`...)
		case '\r':
			buf = append(buf, `\r`...)
		case '\t':
			buf = append(buf, `\t`...)
		case '\v':
			buf = append(buf, `\v`...)
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
	//buf = append(buf, quote)
	return string(buf)

}
