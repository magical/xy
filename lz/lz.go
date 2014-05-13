package lz

import (
	"bufio"
	//"fmt"
	"errors"
	"io"
)

var errMalformed = errors.New("lz.Decode: malformed data")

func decode10(r io.ByteReader, size int) ([]byte, error) {
	var err error
	var nextbyte = func() (b byte) {
		if err == nil {
			b, err = r.ReadByte()
		}
		return
	}
	data := make([]byte, 0, size)
	for len(data) < size && err == nil {
		bits := nextbyte()
		for i := 0; i < 8 && len(data) < size; i, bits = i+1, bits<<1 {
			if bits&0x80 == 0 {
				data = append(data, nextbyte())
				continue
			}
			n := int(nextbyte())<<8 + int(nextbyte())
			count := n>>12 + 3
			disp := n&0xFFF + 1
			if disp > len(data) {
				return nil, errMalformed
			}
			if len(data)+count > size {
				count = size - len(data)
			}
			for j := 0; j < count; j++ {
				data = append(data, data[len(data)-disp])
			}
		}
	}
	if err != nil {
		return nil, err
	}
	return data, nil
}

func decode11(r io.ByteReader, size int) ([]byte, error) {
	var err error
	var nextbyte = func() (b byte) {
		if err == nil {
			b, err = r.ReadByte()
		}
		return
	}
	data := make([]byte, 0, size)
	for len(data) < size && err == nil {
		bits := nextbyte()
		for i := 0; i < 8 && len(data) < size; i, bits = i+1, bits<<1 {
			if bits&0x80 == 0 {
				data = append(data, nextbyte())
				continue
			}
			n := int(nextbyte())<<8 + int(nextbyte())
			count := 1
			switch n >> 12 {
			case 0:
				n = n&0xFFF<<8 + int(nextbyte())
				count = 0x11
			case 1:
				// n doesn't exceed 28 bits
				n = n&0xFFF<<16 + int(nextbyte())<<8 + int(nextbyte())
				count = 0x111
			}
			count += n >> 12
			disp := n&0xFFF + 1
			if disp > len(data) {
				return nil, errMalformed
			}
			if len(data)+count > size {
				count = size - len(data)
			}
			//fmt.Println(len(data), disp, len(data)-disp)
			for j := 0; j < count; j++ {
				data = append(data, data[len(data)-disp])
			}
		}
	}
	if err != nil {
		return nil, err
	}
	return data, nil
}

type sizer interface {
	Size() int64
}

func byteReader(r io.Reader) io.ByteReader {
	if r, ok := r.(io.ByteReader); ok {
		return r
	}
	return bufio.NewReader(r)
}

// Decode expands an LZ-compressed file.
func Decode(r io.Reader) ([]byte, error) {
	var b [4]byte
	_, err := io.ReadFull(r, b[:])
	if err != nil {
		return nil, err
	}
	magic := b[0]
	size := int(b[1]) + int(b[2])<<8 + int(b[3])<<16
	var data []byte
	switch magic {
	case 0x10:
		data, err = decode10(byteReader(r), size)
	case 0x11:
		data, err = decode11(byteReader(r), size)
	default:
		err = errors.New("lz.Decode: not compressed")
	}
	if err == nil && len(data) != size {
		err = errors.New("lz.Decode: size mismatch")
	}
	return data, err
}

// IsCompressed reports whether a reader is likely LZ-compressed. If the reader implements the Size method, it will be used to check the decompressed size against the compressed size.
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
