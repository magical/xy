// Package lz implements the compression algorithm used by the Nintendo 3DS.
package lz

// TODO: Use a sliding window buffer instead of
// storing a copy of the entire decompressed file.

import (
	"bufio"
	"errors"
	"io"

	"fmt"
	"os"
)

const debug = false

var (
	ErrHeader = errors.New("lz: invalid header")
	errMalformed = errors.New("lz: malformed data")
)

type Reader struct {
	reader io.ByteReader
	roffset int
	woffset int
	buf []byte

	err error

	magic byte
	size int
	scratch [4]byte

	decode func() (int, int)

	bits byte
	bitcount int
}

func (z *Reader) readbyte() (b byte) {
	if z.err == nil {
		b, z.err = z.reader.ReadByte()
		if z.err == io.EOF {
			z.err = io.ErrUnexpectedEOF
		}
		z.roffset++
	}
	return
}

func (z *Reader) nextbit() bool {
	if z.bitcount == 0 {
		z.bits = z.readbyte()
		z.bitcount = 8
	}
	bit := z.bits & 0x80 != 0
	z.bitcount--
	z.bits <<= 1
	return bit
}

func (z *Reader) Read(p []byte) (n int, err error) {
	if z.err != nil {
		return 0, z.err
	}

	n = len(z.buf) - z.woffset
	for len(z.buf) < z.size && n < len(p) && z.err == nil {
		if !z.nextbit() {
			z.buf = append(z.buf, z.readbyte())
			n++
			continue
		}

		off := z.roffset
		count, dist := z.decode()
		if dist > len(z.buf) {
			z.err = fmt.Errorf("lz: bad distance %x at %x", dist, off)
			break
		}
		if len(z.buf)+count > z.size {
			z.err = fmt.Errorf("lz: bad size %x (%x remaining) at %x",
				count, z.size - len(z.buf), off)
			count = z.size - len(z.buf)
		}
		//fmt.Println(len(z.buf), dist, len(z.buf)-dist)
		for i := 0; i < count; i++ {
			z.buf = append(z.buf, z.buf[len(z.buf)-dist])
		}
		n += count
	}

	if n < len(p) && z.err == nil {
		z.err = io.EOF
	}
	if n > len(p) {
		n = len(p)
	}
	copy(p, z.buf[z.woffset : z.woffset+n])
	z.woffset += n
	return n, z.err
}

func (z *Reader) decode10() (count, dist int) {
	n := int(z.readbyte())<<8 + int(z.readbyte())
	count = n>>12 + 3
	dist = n&0xFFF + 1
	return
}

func (z *Reader) decode11() (count, dist int) {
	n := int(z.readbyte())<<8 + int(z.readbyte())
	code := n>>12
	switch n >> 12 {
	default:
		// 4-bit count, 12-bit distance
		count = 1
	case 0:
		// 8-bit count, 12-bit distance
		n = n&0xFFF<<8 + int(z.readbyte())
		count = 0x11
	case 1:
		// 16-bit count, 12-bit distance
		// n doesn't exceed 28 bits
		n = n&0xFFF<<16 + int(z.readbyte())<<8 + int(z.readbyte())
		count = 0x111
	}
	count += n >> 12
	dist = n&0xFFF + 1
	if debug {
		fmt.Fprintf(os.Stderr, "code %3d at %x/%x: %x,%x\n",
			code, z.roffset, len(z.buf), count, -dist)
	}
	return
}

func byteReader(r io.Reader) io.ByteReader {
	if r, ok := r.(io.ByteReader); ok {
		return r
	}
	return bufio.NewReader(r)
}

// Decode expands a compressed file.
// It is equivalent to ioutil.ReadAll(NewReader(r)).
func Decode(r io.Reader) ([]byte, error) {
	z, err := NewReader(r)
	if err != nil {
		return nil, err
	}
	data := make([]byte, z.size)
	n, err := z.Read(data)
	return data[:n], err
}

// NewReader returns a new Reader that can be used to read the uncompressed version of r.
func NewReader(r io.Reader) (*Reader, error) {
	z := new(Reader)
	z.reader = byteReader(r)
	b := z.scratch[:]
	_, err := io.ReadFull(r, b)
	if err != nil {
		return nil, err
	}
	z.magic = b[0]
	z.size = int(b[1]) + int(b[2])<<8 + int(b[3])<<16
	switch z.magic {
	case 0x10:
		z.decode = z.decode10
	case 0x11:
		z.decode = z.decode11
	default:
		return nil, ErrHeader
	}
	return z, nil
}

// Size returns the size of the uncompressed data.
func (z *Reader) Size() int64 {
	return int64(z.size)
}

type sizer interface {
	Size() int64
}

// IsCompressed reports whether r looks compressed.
//
// If the reader implements the Size method,
// it will be used to check the decompressed size against the compressed size.
func IsCompressed(r io.Reader) bool {
	var b [4]byte
	n, err := r.Read(b[:])
	if n < 4 || err != nil {
		return false
	}
	if b[0] != 0x10 && b[0] != 0x11 {
		return false
	}
	if r, ok := r.(sizer); ok {
		size := int64(b[1]) + int64(b[2])<<8 + int64(b[3])<<16
		if size < r.Size() {
			return false
		}
	}
	return true
}
