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
	printTree(d.Root, 0)
	for _, f := range d.Files {
		fmt.Println(f.Name)
		io.Copy(hex.Dumper(os.Stdout), f)
		fmt.Println()
		fmt.Println()
	}
}

type indent int

const tabs = "\t\t\t\t\t\t\t\t\t\t\t\t\t\t"

func (i indent) String() string {
	return tabs[len(tabs)-int(i):]
}

func printTree(dir *darc.Dir, indent indent) {
	fmt.Print(indent, dir.Name, "\n")
	for _, f := range dir.Files {
		fmt.Print(indent+1, f.Name, "\n")
	}
	for _, d := range dir.Dirs {
		printTree(d, indent+1)
	}
}
