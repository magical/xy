package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"

	"xy/garc"
	"xy/lz"
)

func usage() {
	fmt.Println("Usage: getfile file.garc number")
	flag.PrintDefaults()
}

func main() {
	expand := flag.Bool("z", false, "attempt to decompress the file")
	flag.Usage = usage
	flag.Parse()
	if flag.NArg() < 2 {
		flag.Usage()
		return
	}
	filename := flag.Arg(0)
	number, _ := strconv.Atoi(flag.Arg(1))
	err := get(filename, number, *expand)
	if err != nil {
		fmt.Println(err)
	}
}

func get(filename string, number int, expand bool) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	files, err := garc.Files(f)
	if err != nil {
		return err
	}

	if number < 0 || number >= len(files) {
		return errors.New("no such file")
	}

	file := files[number]
	var r io.Reader
	if expand {
		r, err = Decode(files[number])
		if err != nil {
			off, _ := file.Seek(0, os.SEEK_CUR)
			fmt.Fprintf(os.Stderr, "%s at %X\n", err, off)
		}
	} else {
		r = file
	}
	io.Copy(os.Stdout, r)
	return nil
}

func Decode(r io.ReadSeeker) (io.Reader, error) {
	data, err := lz.Decode(r)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(data), nil
}
