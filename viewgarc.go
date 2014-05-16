package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"xy/garc"
	"xy/lz"
)

var length int

func main() {
	flag.IntVar(&length, "w", 10, "length of hex dump, in words")
	flag.Parse()
	if flag.NArg() < 1 {
		flag.Usage()
		return
	}
	filename := flag.Arg(0)
	err := view(filename)
	if err != nil {
		fmt.Println(err)
	}
}

func view(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	files, err := garc.Files(f)
	if err != nil {
		return err
	}

	for i, file := range files {
		num := fmt.Sprintf("(%d.%d)", file.Major, file.Minor)
		fmt.Printf("%08x [%8x] %5d %8s: ", file.Offset(), file.Size(), i, num)
		dfile, compressed := tryDecompress(file)
		if compressed {
			fmt.Print("*")
		} else {
			fmt.Print(" ")
		}
		var nn, n int64
		nn, err := hex(dfile, length)
		if err == nil && compressed {
			// Make sure there are no decoding errors
			n, err = io.Copy(ioutil.Discard, dfile)
			nn += n
			if nn != dfile.Size() {
				fmt.Printf("size mismatch: %d expected %d\n", nn, dfile.Size())
			}
		}
		if err != nil {
			fmt.Println(err)
		}
	}
	return nil
}

type readerSize interface {
	io.Reader
	Size() int64
}

func tryDecompress(r *garc.File) (readerSize, bool) {
	z, err := lz.NewReader(r)
	if err != nil || z.Size() < r.Size() {
		r.Seek(0, 0)
		return r, false
	}
	return z, true
}

func hex(r io.Reader, length int) (nn int64, err error) {
	const hex = "0123456789ABCDEF"
	var b [4]byte
	var buf bytes.Buffer
	for {
		n, err := r.Read(b[:])
		nn += int64(n)
		for i := 0; i < n; i++ {
			buf.WriteByte(hex[b[i]>>4])
			buf.WriteByte(hex[b[i]&0xF])
		}
		if buf.Len() >= length*9-1 {
			buf.WriteString("...")
			break
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return nn, err
		}
		buf.WriteByte(' ')
	}
	buf.WriteByte('\n')
	io.Copy(os.Stdout, &buf)
	return nn, nil
}
