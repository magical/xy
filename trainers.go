package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"garc"
	"log"
	"os"
	"path/filepath"
	"xy/names"
)

const (
	trdataPath = "a/0/3/8"
	trpokePath = "a/0/4/0"
)

type Trdata struct {
	HasMoves     uint8
	TrainerClass uint8
	BattleType   uint8
	NumPokemon   uint8
	Items        [4]uint8
	_            uint32
	_            uint32
}

type Trpoke struct {
	Unknown uint16
	Level   uint16
	Pokemon uint16
	Form    uint16
	Moves   []uint16
	Item    uint16
}

type Trpoke0 struct {
	Unknown uint16
	Level   uint16
	Pokemon uint16
	Form    uint16
}

type Trpoke1 struct {
	Trpoke0
	Moves [4]uint16
}

type Trpoke2 struct {
	Trpoke0
	Item uint16
}

type Trpoke3 struct {
	Trpoke0
	_ [7]uint32
}

func parse_trdata(f *garc.File) (Trdata, error) {
	var v Trdata
	if err := binary.Read(f, binary.LittleEndian, &v); err != nil {
		return Trdata{}, err
	}
	return v, nil
}

func parse_trpoke(info *Trdata, f *garc.File) ([]Trpoke, error) {
	v := make([]Trpoke, info.NumPokemon)
	if info.HasMoves == 0 {
		n := f.Size() / 8
		v0 := make([]Trpoke0, n)
		if err := binary.Read(f, binary.LittleEndian, v0); err != nil {
			return nil, err
		}
		for i, p := range v0 {
			v[i].Unknown = p.Unknown
			v[i].Level = p.Level
			v[i].Pokemon = p.Pokemon
			v[i].Form = p.Form
		}
		return v, nil
	} else if info.HasMoves == 1 {
		n := f.Size() / 16
		v0 := make([]Trpoke1, n)
		if err := binary.Read(f, binary.LittleEndian, v0); err != nil {
			return nil, err
		}
		for i, p := range v0 {
			v[i].Unknown = p.Unknown
			v[i].Level = p.Level
			v[i].Pokemon = p.Pokemon
			v[i].Form = p.Form
			for j := range p.Moves {
				if p.Moves[j] != 0 {
					v[i].Moves = append(v[i].Moves, p.Moves[j])
				}
			}
		}
		return v, nil
	} else if info.HasMoves == 2 {
		n := f.Size() / 10
		v0 := make([]Trpoke2, n)
		if err := binary.Read(f, binary.LittleEndian, v0); err != nil {
			return nil, err
		}
		for i, p := range v0 {
			v[i].Unknown = p.Unknown
			v[i].Level = p.Level
			v[i].Pokemon = p.Pokemon
			v[i].Form = p.Form
			v[i].Item = p.Item
		}
		return v, nil
	} else if info.HasMoves == 3 {
		n := f.Size() / 36
		v0 := make([]Trpoke3, n)
		if err := binary.Read(f, binary.LittleEndian, v0); err != nil {
			return nil, err
		}
		for i, p := range v0 {
			v[i].Unknown = p.Unknown
			v[i].Level = p.Level
			v[i].Pokemon = p.Pokemon
			v[i].Form = p.Form
		}
		return v, nil
	} else {
		return nil, fmt.Errorf("unknown trainer type: %d", info.HasMoves)
	}
}

func (p Trpoke) String() string {
	species := int(p.Pokemon)
	return fmt.Sprintf("L%d %s %d (%x)", p.Level, names.Species(species), p.Form, p.Unknown)
}

func main() {
	flag.Parse()
	if err := main1(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func loadTrdata(filename string) ([]Trdata, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	files, err := garc.Files(f)
	if err != nil {
		return nil, err
	}
	var trdata = make([]Trdata, len(files))
	for i, f := range files {
		trdata[i], err = parse_trdata(f)
		if err != nil {
			if i != 0 {
				fmt.Println(err)
			}
			continue
		}
	}
	return trdata, nil
}

func main1() error {
	romfsPath := flag.Arg(0)

	trdata, err := loadTrdata(filepath.Join(romfsPath, filepath.FromSlash(trdataPath)))
	if err != nil {
		return err
	}

	trpoke, err := os.Open(filepath.Join(romfsPath, filepath.FromSlash(trpokePath)))
	if err != nil {
		return err
	}

	files, err := garc.Files(trpoke)
	if err != nil {
		return err
	}
	for i, f := range files {
		pokes, err := parse_trpoke(&trdata[i], f)
		if err != nil {
			log.Println(err)
			continue
		}
		fmt.Println(i, names.TrainerClass(int(trdata[i].TrainerClass)), names.TrainerName(i))
		for _, p := range pokes {
			fmt.Println(i, "-", p)
		}
		fmt.Println()
	}
	return nil
}
