package util

import (
	"os"
	"xy/garc"
)

type GARC struct {
	Files []*garc.File
	f *os.File
}

func (g *GARC) Close() error {
	return g.f.Close()
}

func OpenGARC(name string) (*GARC, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	files, err := garc.Files(f)
	if err != nil {
		return nil, err
	}
	return &GARC{Files: files, f: f}, nil
}
