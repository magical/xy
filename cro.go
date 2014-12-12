package main

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/kr/pretty"
)

var le = binary.LittleEndian

var Magic = [4]byte{'C', 'R', 'O', '0'}

type Header struct {
	HashTable [0x80]uint8
	Magic     [4]byte
	CodeSize0 uint32
	A         [2]uint32
	FileSize  uint32
	B         [6]uint32
	BaseAddr  uint32

	CodeOffset    uint32
	CodeSize      uint32
	C            uint32
	D             uint32
	NameOffset    uint32
	NameSize      uint32
	SegmentOffset uint32
	SegmentCount  uint32

	ExportOffset       uint32
	ExportCount        uint32
	E                  uint32
	F                  uint32
	ExportStringOffset uint32
	ExportStringSize   uint32
	ExportTreeOffset   uint32
	ExportTreeSize     uint32
	G                  uint32
	H                  uint32

	ImportPatchOffset uint32
	ImportPatchCount  uint32
	ImportTable       [3]struct {
		Offset uint32
		Count  uint32
	}
	ImportStringOffset uint32
	ImportStringSize   uint32
	I                  [2]uint32
	PatchOffset        uint32
	PatchCount         uint32
	J                  [2]uint32

	K [18]uint32
}

type Segment struct {
	Offset uint32
	Size uint32
	ID   uint32
}

type Patch struct {
	Dest uint32
	Type uint8
	Seg uint8
	_ uint8
	_ uint8
	X uint32
}

type Import struct {
	NameOffset   uint32
	SymbolOffset uint32
}

type Export struct {
	NameOffset uint32
	DataOffset uint32
}

func main() {
	dump := flag.Bool("dump", false, "")
	flag.Parse()
	if *dump {
		dumpmain()
	} else {
		printmain()
	}
}

func dumpmain() {
	f, err := os.Open(flag.Arg(0))
	if err != nil {
		fmt.Println(err)
		return
	}
	var header Header
	err = binary.Read(f, le, &header)
	if err != nil {
		fmt.Println(err)
		return
	}
	if header.Magic != Magic {
		fmt.Println("not a cro")
		return
	}

	segments := make([]Segment, header.SegmentCount)
	_, err = f.Seek(int64(header.SegmentOffset), 0)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = binary.Read(f, le, segments)
	if err != nil {
		fmt.Println(err)
		return
	}

	seg := segments[0]
	_, err = f.Seek(int64(seg.Offset), 0)
	if err != nil {
		fmt.Println(err)
		return
	}
	_, err = io.CopyN(os.Stdout, f, int64(seg.Size))
	if err != nil {
		fmt.Println(err)
		return
	}
}

func printmain() {
	f, err := os.Open(flag.Arg(0))
	if err != nil {
		fmt.Println(err)
		return
	}
	contents, err := ioutil.ReadFile(flag.Arg(0))
	if err != nil {
		fmt.Println(err)
		return
	}
	var header Header
	err = binary.Read(f, le, &header)
	if err != nil {
		fmt.Println(err)
		return
	}
	if header.Magic != Magic {
		fmt.Println("not a cro")
		return
	}
	pretty.Println(header)

	fmt.Printf("Name: %s\n", contents[header.NameOffset:][:header.NameSize-1])

	fmt.Println("Segments")
	segments := make([]Segment, header.SegmentCount)
	f.Seek(int64(header.SegmentOffset), 0)
	err = binary.Read(f, le, segments)
	if err != nil {
		fmt.Println(err)
		return
	}
	pretty.Println(segments)

	fmt.Println("Exports")
	exports := make([]Export, header.ExportCount)
	f.Seek(int64(header.ExportOffset), 0)
	err = binary.Read(f, le, exports)
	if err != nil {
		fmt.Println(err)
		return
	}
	//fmt.Println(exports)
	for _, e := range exports {
		name := contents[e.NameOffset:]
		name = name[:bytes.IndexByte(name, 0)]
		fmt.Printf("%x %s\n", e.DataOffset, name)
	}

	err =patch(f, &header, segments, contents)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Print(hex.Dump(contents[segments[0].Offset:][:segments[0].Size]))

	if false {
		for _, imp := range header.ImportTable {
			imports := make([]Import, imp.Count)

			f.Seek(int64(imp.Offset), 0)
			err = binary.Read(f, le, imports)
			if err != nil {
				fmt.Println(err)
				return
			}
			for i, imp := range imports {
				fmt.Printf("Import %d\n", i)
				fmt.Printf(" Name offset: %x\n", imp.NameOffset)
				fmt.Printf(" Symbol offset: %x\n", imp.SymbolOffset)
				/*
				var name []byte
				if int(imp.NameOffset) < len(contents) {
					name = contents[imp.NameOffset:]
					name = name[:bytes.IndexByte(name, 0)]
				}
				fmt.Printf("%x %s\n", imp.SymbolOffset, name)
				*/
			}
		}
	}
}

func patch(f *os.File, header *Header, segments []Segment, contents []byte)  error {
	// Patches
	patches := make([]Patch, header.PatchCount)
	f.Seek(int64(header.PatchOffset), 0)
	err := binary.Read(f, le, patches)
	if err != nil {
		return err
	}

	for _, p := range patches {
		dest := p.Dest >> 4
		seg := int(p.Dest&0xF)
		if seg > len(segments) {
			fmt.Fprintln(os.Stderr, "segment out of range")
			continue
		}
		base := segments[seg].Offset
		if int(p.Seg) > len(segments) {
			fmt.Fprintln(os.Stderr, "segment out of range")
			continue
		}
		xbase := segments[p.Seg].Offset
		switch p.Type {
		case 2:
			// Absolute address
			off := base + dest
			le.PutUint32(contents[off:], xbase + p.X)
		}
	}
	return nil
}
