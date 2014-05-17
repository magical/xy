// Usage: extractdarc garcfile outdir
// Extracts all darc files in garcfile to outdir.
package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	pathlib "path"
	"path/filepath"

	"xy/darc"
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

	for _, gfile := range gfiles {
		var ff darc.Reader
		data, err := lz.Decode(gfile)
		if err == nil {
			ff = bytes.NewReader(data)
		} else {
			gfile.Seek(0, 0)
			ff = gfile
		}
		d, err := darc.Read(ff)
		if err != nil {
			continue
		}
		garcname := fmt.Sprintf("%d.%d", gfile.Major, gfile.Minor)
		os.Mkdir(outdir, 0777)
		err = os.Mkdir(filepath.Join(outdir, garcname), 0777)
		if err != nil && !os.IsExist(err) {
			log.Println(err)
			return
		}
		walk(d, func(path string, file *darc.File) {
			errname := pathlib.Join(filename+"["+garcname+"]", path, file.Name)
			outname := filepath.Join(outdir, garcname, path, file.Name)
			os.Mkdir(filepath.Join(outdir, garcname, path), 0777)
			//log.Println(outname)
			//_  =errname
			out, err := os.Create(outname)
			if err != nil {
				log.Printf("%s: %s", errname, err)
				return
			}
			_, err = io.Copy(out, file)
			if err != nil {
				log.Printf("%s: %s", errname, err)
			}
		})
	}
}

func walk(d *darc.DARC, callback func(path string, file *darc.File)) {
	_walk(d.Root, "", callback)
}

func _walk(root *darc.Dir, path string, callback func(path string, file *darc.File)) {
	for _, f := range root.Files {
		callback(path, f)
	}
	for _, d := range root.Dirs {
		path := pathlib.Join(path, d.Name)
		_walk(d, path, callback)
	}
}
