package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"

	"xy/darc"
)

func main() {
	var in darc.Reader
	if len(os.Args) > 1 {
		f, err := os.Open(os.Args[1])
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		in = f
	} else {
		data, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			log.Fatal(err)
		}
		in = bytes.NewReader(data)
	}
	d, err := darc.Read(in)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%+v\n", d)
	for _, f := range d.Files {
		fmt.Println(f.Name)
		io.Copy(hex.Dumper(os.Stdout), f)
		fmt.Println()
		fmt.Println()
	}
}
