package garc

import (
	"encoding/binary"
	"errors"
	"io"
)

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
	// Followed by RecordCount words, each an offset into the FATB data.
	// The first word is a bit vector. For each set bit, a Record follows.
}

// File allocation table
type FATB struct {
	Magic       [4]byte // BTAF
	Size        uint32
	RecordCount uint32
}

type Record struct {
	Start uint32
	End   uint32
	Size  uint32
}

type Reader interface {
	io.Reader
	io.ReaderAt
	io.Seeker
}

type File struct {
	io.SectionReader
	off int64
}

func (f *File) Offset() int64 {
	return f.off
}

func Files(r Reader) ([]*File, error) {
	var head Header
	var fato FATO
	var fatb FATB

	err := binary.Read(r, binary.LittleEndian, &head)
	if err != nil {
		return nil, err
	}
	//fmt.Println(head)
	if string(head.Magic[:]) != "CRAG" {
		return nil, errors.New("not a GRAC")
	}

	err = binary.Read(r, binary.LittleEndian, &fato)
	if err != nil {
		return nil, err
	}

	osets := make([]uint32, fato.RecordCount)
	err = binary.Read(r, binary.LittleEndian, &osets)
	if err != nil {
		return nil, err
	}

	err = binary.Read(r, binary.LittleEndian, &fatb)
	if err != nil {
		return nil, err
	}

	files := make([]*File, 0, fatb.RecordCount)
	for _ = range osets {
		var vec uint32
		err := binary.Read(r, binary.LittleEndian, &vec)
		if err != nil {
			return nil, err
		}
		var rec Record
		for ; vec != 0; vec >>= 1 {
			if vec&1 == 0 {
				continue
			}
			err = binary.Read(r, binary.LittleEndian, &rec)
			if err != nil {
				return nil, err
			}
			off := int64(head.DataOffset) + int64(rec.Start)
			size := int64(rec.Size)
			files = append(files, &File{*io.NewSectionReader(r, off, size), off})
		}
	}
	return files, nil
}
