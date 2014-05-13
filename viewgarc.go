package main

import (
	"bytes"
	"os"
	"io"
	"errors"
	"encoding/binary"
	"fmt"
	"flag"
)

type GRAC struct {
}

type Header struct {
	Magic [4]byte
	HeaderSize uint32
	BOM uint32
	ChunkCount uint32 // always 4
	DataOffset uint32
	Size uint32
	LastSize uint32 // same as last word in FATB
}

// File allocation table offsets
type FATO struct {
	Magic [4]byte // OTAF
	Size uint32
	RecordCount uint16
	_ uint16 // always 0xFFFF
	// Each record is 4 bytes and gives an offset into the FATB
}

// File allocation table
type FATB struct {
	Magic [4]byte // BTAF
	Size uint32
	RecordCount uint32
}

type Record struct {
	_ uint32 // always 1
	Start uint32
	End uint32
	Size uint32
}

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

	var head Header
	var fato FATO
	var fatb FATB

	err = binary.Read(f, binary.LittleEndian, &head)
	if err != nil {
		return err
	}
	//fmt.Println(head)
	if string(head.Magic[:]) != "CRAG" {
		return errors.New("not a GRAC")
	}

	err = binary.Read(f, binary.LittleEndian, &fato)
	if err != nil {
		return err
	}

	// Skip FATO
	f.Seek(int64(fato.Size) - 12, 1)

	err = binary.Read(f, binary.LittleEndian, &fatb)
	if err != nil {
		return err
	}

	records := make([]Record, fatb.RecordCount)
	err = binary.Read(f, binary.LittleEndian, records)
	if err != nil {
		return err
	}

	for i, rec := range records {
		off := int64(head.DataOffset) + int64(rec.Start)
		size := int64(rec.Size)
		r := io.NewSectionReader(f, off, size)
		fmt.Printf("%08x [%8x] %8d: ", off, size, i)
		err = hex(r, length)
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
