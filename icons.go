package main

import (
	"bytes"
	"errors"
	"encoding/binary"
	"fmt"
	"image"
	"image/png"
	"image/color"
	"io"
	"log"
	"os"
	"path/filepath"

	"xy/garc"
	"xy/lz"
)

var le = binary.LittleEndian

type RGB15 uint16

func (c RGB15) NRGBA() color.NRGBA {
	return color.NRGBA{
		R: uint8((uint32(c>>11&31)*0xFF + 15) / 31),
		G: uint8((uint32(c>>6&31)*0xFF + 15) / 31),
		B: uint8((uint32(c>>1&31)*0xFF + 15) / 31),
		A: uint8(c&1)*0xFF,
	}
}

func main() {
	filename := os.Args[1]
	outdir := os.Args[2]
	f, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
		return
	}

	gfiles, err := garc.Files(f)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, file := range gfiles {
		errname := fmt.Sprintf("%s[%d.%d]", filename, file.Major, file.Minor)
		m, err := decode(file)
		if err != nil {
			log.Printf("%s: %s", errname, err)
			continue
		}
		out, err := os.Create(filepath.Join(outdir, fmt.Sprintf("%d.png", file.Major)))
		if err != nil {
			log.Printf("%s: %s", err)
			return
		}
		png.Encode(out, m)
	}
}

func decode(f *garc.File) (*image.Paletted, error) {
	z, err := lz.Decode(f)
	if err != nil {
		return nil, err
	}
	r := bytes.NewReader(z)
	r.Seek(-0x14, os.SEEK_END)

	var imagHeader struct{
		Magic [4]byte
		HeaderSize uint32
		Width uint16
		Height uint16
		_ uint32
		DataSize uint32
	}
	err = binary.Read(r, le, &imagHeader)
	if err != nil {
		return nil, err
	}
	if string(imagHeader.Magic[:]) != "imag" {
		return nil, errors.New("not an imag")
	}

	w := int(imagHeader.Width)
	h := int(imagHeader.Height)

	w = (w+31) &^ 31
	h = (h+31) &^ 31

	r.Seek(0, os.SEEK_SET)
	var paletteHdr [2]uint16
	err = binary.Read(r, le, &paletteHdr)
	if err != nil {
		return nil, err
	}
	paletteCount := paletteHdr[1]
	colors := make([]RGB15, paletteCount)
	err = binary.Read(r, le, colors)
	if err != nil {
		return nil, err
	}

	pal := make(color.Palette, paletteCount)
	for i, c := range colors {
		pal[i] = c.NRGBA()
	}

	rect := image.Rect(0, 0, int(imagHeader.Width), int(imagHeader.Height))
	m := image.NewPaletted(rect, pal)
	// os.Stdin.Read(m.Pix)

	var buf = make([]byte, int(imagHeader.DataSize) - 4 - len(pal)*2)
	_, err = io.ReadFull(r, buf)
	if err != nil {
		return nil, err
	}
	ti := 0
	const T = 8
	var tileBuf [T*T]uint8
	var tile []uint8
	for y := 0; y < h; y += T {
		for x := 0; x < w; x += T {
			if paletteCount <= 16 {
				for i := 0; i < T*T; i += 2 {
					p := buf[ti*T*T/2 + i/2]
					tileBuf[i] = p>>4
					tileBuf[i+1] = p&0xF
				}
				tile = tileBuf[:]
			} else {
				tile = buf[ti*T*T:]
			}
			for ty := 0; ty < T; ty++ {
				for tx := 0; tx < T; tx++ {
					si := mingle(ty, tx)
					m.SetColorIndex(x+tx, y+ty, tile[si])
				}
			}
			ti++
		}
	}
	return m, nil
}

// Mingle interleaves the lower 4 bits of x and y
func mingle(x, y int) int {
	x = (x | x<<2) & 0x33
	x = (x | x<<1) & 0x55
	y = (y | y<<2) & 0x33
	y = (y | y<<1) & 0x55
	return x<<1 | y
}
