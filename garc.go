package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
)

type GARC struct {
}

type Header struct {
	Magic      [4]byte
	HeaderSize uint32
	BOM        uint32
	ChunkCount uint32 // always 4
	DataOffset uint32
	Size       uint32
	LastSize   uint32 // same as last word in FATB
}

// File allocation table offsets
type FATO struct {
	Magic       [4]byte // OTAF
	Size        uint32
	RecordCount uint16
	_           uint16 // always 0xFFFF
	// Each record is 4 bytes and gives an offset into the FATB
}

// File allocation table
type FATB struct {
	Magic       [4]byte // BTAF
	Size        uint32
	RecordCount uint32
}

type Record struct {
	_     uint32 // always 1
	Start uint32
	End   uint32
	Size  uint32
}

func main() {
	for _, filename := range os.Args[1:] {
		if st, err := os.Stat(filename); err == nil && st.Mode().IsDir() {
			continue
		}
		err := do(filename)
		if err != nil {
			st, _ := os.Stat(filename)
			fmt.Printf("%s: %8x %s\n", filename, st.Size(), err)
		}
	}
}

func do(filename string) error {
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
		return errors.New("not a GARC")
	}

	err = binary.Read(f, binary.LittleEndian, &fato)
	if err != nil {
		return err
	}

	// Skip FATO
	f.Seek(int64(fato.Size)-12, 1)

	err = binary.Read(f, binary.LittleEndian, &fatb)
	if err != nil {
		return err
	}

	rec := make([]Record, fatb.RecordCount)
	err = binary.Read(f, binary.LittleEndian, rec)
	if err != nil {
		return err
	}

	fmt.Printf("%s: %8x %5d\n", filename, head.Size, len(rec))
	return nil
}
