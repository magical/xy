package main

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/png"
	"image/color"
	"io"
	"os"
	"strconv"
)

var le = binary.LittleEndian

type RGBA4444 uint16

func (c RGBA4444) RGBA() (r, g, b, a uint32) {
	const m = 0xFFFF
	r = (uint32(c>>12&15)*m + 7) / 15
	g = (uint32(c>>8&15)*m + 7) / 15
	b = (uint32(c>>4&15)*m + 7) / 15
	a = (uint32(c>>0&15)*m + 7) / 15
	return r*a/m, g*a/m, b*a/m, a
}

type RGB15 uint16

func (c RGB15) RGBA() (r, g, b, a uint32) {
	const m = 0xFFFF
	r = (uint32(c>>11&31)*m + 15) / 31
	g = (uint32(c>>6&31)*m + 15) / 31
	b = (uint32(c>>1&31)*m + 15) / 31
	a = m
	return r, g, b, a
}

func main() {
	w, _ := strconv.Atoi(os.Args[1])
	h, _ := strconv.Atoi(os.Args[2])

	r := os.Stdin

	var paletteHdr [2]uint16
	err := binary.Read(r, le, &paletteHdr)
	if err != nil {
		fmt.Println(err)
		return
	}
	paletteCount := paletteHdr[1]
	palette := make([]RGB15, paletteCount)
	err = binary.Read(r, le, palette)
	if err != nil {
		fmt.Println(err)
		return
	}

	pal := make(color.Palette, paletteCount)
	for i, c := range palette {
		if i == 0 {
			n := color.NRGBAModel.Convert(c).(color.NRGBA)
			//n.A = 0
			pal[i] = n
		} else {
			pal[i] = c
		}
	}

	rect := image.Rect(0, 0, w, h)
	m := image.NewPaletted(rect, pal)
	// os.Stdin.Read(m.Pix)

	var buf = make([]byte, w*h/2)
	_, err = io.ReadFull(r, buf)
	if err != nil {
		println(err.Error())
		return
	}
	ti := 0
	const T = 8
	var tile [T*T]uint8
	for y := 0; y < h; y += T {
		for x := 0; x < w; x += T {
			for i := 0; i < T*T; i += 2 {
				p := buf[ti*T*T/2 + i/2]
				tile[i] = p>>4
				tile[i+1] = p&0xF
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
	png.Encode(os.Stdout, m)
}

// Mingle interleaves the lower 4 bits of x and y
func mingle(x, y int) int {
	x = (x | x<<2) & 0x33
	x = (x | x<<1) & 0x55
	y = (y | y<<2) & 0x33
	y = (y | y<<1) & 0x55
	return x<<1 | y
}
