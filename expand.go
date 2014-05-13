
package main

import (
	"bytes"
	"os"
	"io"
	"io/ioutil"
	"fmt"
	"flag"

	"github.com/mattn/go-isatty"

	"xy/lz"
)

func usage() {
	fmt.Println("Usage: expand [input] >output")
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	flag.Parse()
	filename := flag.Arg(0)
	err := expand(filename)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func expand(filename string) error {
	in := os.Stdin
	var out io.Writer = os.Stdout
	if filename != "" {
		var err error
		in, err = os.Open(filename)
		if err != nil {
			return err
		}
		defer in.Close()
	}

	if isatty.IsTerminal(os.Stdout.Fd()) {
		fmt.Fprintln(os.Stderr, "discarding output instead of sending to terminal")
		out = ioutil.Discard
	}

	r, err := Decode(in)
	if r != nil {
		io.Copy(out, r)
	}
	if err != nil {
		off, _ := in.Seek(0, os.SEEK_CUR)
		err = fmt.Errorf("%s at %X", err, off)
	}
	return err
}

func Decode(r io.ReadSeeker) (io.Reader, error) {
	data, err := lz.Decode(r)
	if data != nil {
		return bytes.NewReader(data), nil
	}
	return nil, err
}
