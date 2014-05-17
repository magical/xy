// Usage: extractgarc garcfile outdir
// Extract and decompress all files in garcfile to outdir.
package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"xy/garc"
	"xy/lz"
)

func main() {
	filename := os.Args[1]
	outdir := os.Args[2]
	if filename == "" || outdir == "" {
		return
	}
	f, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	gfiles, err := garc.Files(f)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, file := range gfiles {
		z, err := lz.NewReader(file)
		if err != nil {
			//log.Println(err)
			continue
		}
		outname := fmt.Sprintf("%d.%d", file.Major, file.Minor)
		out, err := os.Create(filepath.Join(outdir, outname))
		if err != nil {
			log.Printf("%s[%s]: %s", filename, outname, err)
			break
		}
		_, err = io.Copy(out, z)
		if err != nil {
			log.Printf("%s[%s]: %s", filename, outname, err)
		}
	}
}
