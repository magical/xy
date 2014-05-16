package darc

import (
	"bytes"
	"encoding/binary"
	"errors"
	"github.com/bradfitz/iter"
	"io"
	"unicode/utf16"
)

var ErrHeader = errors.New("darc: invalid header")

type DARC struct {
	Header Header

	Filenames []string
	Uhh       []string

	Root *Dir
	Files []*File
}

type Header struct {
	Magic       [4]byte // "darc"
	BOM         uint16  // Always 0xFEFF
	HeaderSize  uint16  // Always 0x1C
	_           uint32  // Always 1<<24
	Size        uint32  // Size from "darc" to end of file
	RecordOffset uint32  // Always 0x1c
	RecordSize   uint32  // e.g. 0x522
	DataOffset  uint32  // e.g. 0x540
}

type Dir struct {
	Name string
	Parent *Dir
	Dirs []*Dir
	Files []*File
}

type File struct {
	Name string
	*io.SectionReader
}

func (f *File) String() string {
	return f.Name
}

var le = binary.LittleEndian

const (
	seekSet = 0
	seekCur = 1
)

type Reader interface {
	io.Reader
	io.ReaderAt
	io.Seeker
}

func Read(r Reader) (darc *DARC, err error) {
	var num uint32
	darc = new(DARC)

	err = binary.Read(r, le, &num)
	if err != nil { return nil, ErrHeader }
	_, err = r.Seek(int64(num)*64, seekCur)
	if err != nil { return nil, ErrHeader }
	err = binary.Read(r, le, &num)
	if err != nil { return nil, ErrHeader }
	_, err = r.Seek(int64(num)*32, seekCur)
	if err != nil { return nil, ErrHeader }

	/*
	err = binary.Read(r, le, &num)
	if err != nil {
		return
	}
	darc.Filenames, err = readStrings(r, num, 0x40)
	if err != nil {
		return
	}

	err = binary.Read(r, le, &num)
	if err != nil {
		return
	}
	darc.Uhh, err = readStrings(r, num, 0x20)
	if err != nil {
		return
	}
	*/

	off, err := r.Seek(0, seekCur)
	if err != nil {
		return nil, ErrHeader
	}
	off = (off + 0x7F) &^ 0x7F
	r.Seek(off, seekSet)

	h := &darc.Header
	err = binary.Read(r, le, h)
	if err != nil {
		return nil, ErrHeader
	}

	if string(h.Magic[:]) != "darc" {
		return nil, ErrHeader
	}

	rr := io.NewSectionReader(r,
		off + int64(h.RecordOffset),
		off + int64(h.RecordOffset) + int64(h.RecordSize),
	)

	// Read file tree

	var rootRecord [3]uint32
	err = binary.Read(rr, le, &rootRecord)
	if err != nil {
		return
	}

	var records = make([][3]uint32, rootRecord[2])
	records[0] = rootRecord
	err = binary.Read(rr, le, records[1:])
	if err != nil {
		return
	}

	remainingBytes := int(h.RecordSize) - 12*len(records)
	names := make([]uint16, remainingBytes/2)
	err = binary.Read(rr, le, names)
	if err != nil {
		return
	}

	/*
	for _, rec := range records {
		name := decode(names, rec[0])
		var file *File
		var dir *Dir
		if rec[0] >> 24 == 0 {
			rr := io.NewSectionReader(r, off+int64(rec[1]), off+int64(rec[1]+rec[2]))
			file := &File{name, rr}
			darc.Files = append(darc.Files, file)
		}
	}
	*/

	root := &Dir{Name: decode(names, rootRecord[0])}
	root.Parent = root
	buildTree(darc, r, off, root, names, records, 1)

	darc.Root = root

	return
}

func buildTree(darc *DARC, r Reader, off int64, parent *Dir, names []uint16, rec [][3]uint32, i int) int {
	name := decode(names, rec[i][0])
	if rec[i][0] >> 24 == 0 {
		rr := io.NewSectionReader(r,
			off+int64(rec[i][1]),
			off+int64(rec[i][1]) + int64(rec[i][2]),
		)
		file := &File{name, rr}
		if parent != nil {
			parent.Files = append(parent.Files, file)
		}
		darc.Files = append(darc.Files, file)
		return i+1
	}
	dir := &Dir{Name: name, Parent: parent}
	var j = i+1
	for j < int(rec[i][2]) {
		j = buildTree(darc, r, off, dir, names, rec, j)
	}
	parent.Dirs = append(parent.Dirs, dir)
	return j
}

func readStrings(r io.Reader, n uint32, size int) (s []string, err error) {
	// sanity check
	if n > 1000 {
		return nil, errors.New("darc: too many strings")
	}
	b := make([]byte, int(n)*size)
	s = make([]string, 0, n)
	err = binary.Read(r, le, b)
	if err != nil {
		return nil, err
	}
	for i := range iter.N(int(n)) {
		b := b[size*i : size*(i+1)]
		b = bytes.TrimRight(b, "\x00")
		s = append(s, string(b))
	}
	return s, nil
}

func decode(b []uint16, i uint32) string {
	i &= 0xFFFFFF
	if int(i/2) > len(b) {
		return ""
	}
	b = b[i/2:]
	for j := range b {
		if b[j] == 0 {
			b = b[:j]
			break
		}
	}
	return string(utf16.Decode(b))
}
