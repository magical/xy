package main

import (
	"bytes"
	"fmt"
	//"io"
	"io/ioutil"
	"os"
	"strings"

	"xy/darc"
	"xy/garc"
	"xy/lz"
)

func main() {
	for _, filename := range os.Args[1:] {
		if st, err := os.Stat(filename); err == nil && st.Mode().IsDir() {
			continue
		}
		err := do(filename)
		if err != nil {
			fmt.Printf("%s: %s\n", filename, err)
		}
	}
}

func do(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	gfiles, err := garc.Files(f)
	if err != nil {
		return err
	}

	for _, gfile := range gfiles {
		var gf darc.Reader = gfile
		if z, err := lz.NewReader(gfile); err == nil{
			data, err := ioutil.ReadAll(z)
			if err == nil {
				gf = bytes.NewReader(data)
			}
		}
		gfile.Seek(0, 0)
		d, err := darc.Read(gf)
		if err == darc.ErrHeader {
			continue
		}
		if err != nil {
			fmt.Printf("%s[%d.%d]: %s\n", filename, gfile.Major, gfile.Minor, err)
			continue
		}
		fmt.Printf("%s[%d.%d]\n", filename, gfile.Major, gfile.Minor)
		printTree(d.Root.Dirs[0], "")
		fmt.Println()
	}
	return nil
}

func printTree(root *darc.Dir, prefix string) {
	if root.Name != "." {
		fmt.Printf("%s%s\n", prefix, root.Name)
	}
	for _, d := range root.Dirs {
		printTree(d, prefix+"    ")
	}
	for _, f := range root.Files {
		fprefix := prefix + "    "
		fmt.Printf("%s%s\n", fprefix, f.Name)
	}
}

var _spaces = strings.Repeat(" ", 80)
func spaces(n int) string{
	if n < len(_spaces) {
		return _spaces[len(_spaces) - n:]
	}
	return strings.Repeat(" ", n)
}
