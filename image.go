package main

import (
	"image"
	"image/png"
	"io"
	"os"
	"strconv"
)

type RGBA4444 uint16

func (c RGBA4444) RGBA() (r, g, b, a uint32) {
	const m = 0xFFFF
	r = (uint32(c>>12&15)*m + 7) / 15
	g = (uint32(c>>8&15)*m + 7) / 15
	b = (uint32(c>>4&15)*m + 7) / 15
	a = (uint32(c>>0&15)*m + 7) / 15
	return r*a/m, g*a/m, b*a/m, a
}

func main() {
	w, _ := strconv.Atoi(os.Args[1])
	h, _ := strconv.Atoi(os.Args[2])
	r := image.Rect(0, 0, w, h)
	m := image.NewNRGBA(r)
	// os.Stdin.Read(m.Pix)

	bpp := 2 // bytes per pixel

	var buf = make([]byte, w*h*bpp)
	_, err := io.ReadFull(os.Stdin, buf)
	if err != nil {
		println(err.Error())
		return
	}
	ti := 0
	const T = 8
	for x := 0; x < w; x += T {
		for y := 0; y < h; y += T {
			tile := buf[ti*T*T*bpp:]
			for tx := 0; tx < T; tx++ {
				for ty := 0; ty < T; ty++ {
					si := mingle(tx, ty) * bpp
					c := RGBA4444(tile[si]) |
						RGBA4444(tile[si+1])<<8
					m.Set(x+tx, h-y-ty-1, c)
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
