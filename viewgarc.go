package main

import (
	"bytes"
	"os"
	"io"
	"fmt"
	"flag"

	"xy/garc"
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
		fmt.Printf("%08x [%8x] %8d: ", file.Offset(), file.Size(), i)
		err = hex(file, length)
		if err != nil {
			fmt.Printf("%s\n", err)
		}
	}
	return nil
}

func hex(r io.Reader, length int) error{
	const hex = "0123456789ABCDEF"
	var b[4]byte
	var buf bytes.Buffer
	for {
		n, err := r.Read(b[:])
		for i := 0; i < n; i++ {
			buf.WriteByte(hex[b[i]>>4])
			buf.WriteByte(hex[b[i]&0xF])
		}
		if buf.Len() >= length*9 - 1 {
			buf.WriteString("...")
			break
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		buf.WriteByte(' ')
	}
	buf.WriteByte('\n')
	io.Copy(os.Stdout, &buf)
	return nil
}
