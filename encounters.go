package main

import (
	"bytes"
	"database/sql"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"xy/garc"
	"xy/lz"
	"xy/names"

	_ "github.com/lib/pq"
)

const VersionID = 24
const VersionGroupID = 15

var le = binary.LittleEndian

type Encounter struct {
	Header [16]byte
	Grass  [12]Slot
	Flower [3][12]Slot
	Rough  [12]Slot

	Water     [5]Slot
	RockSmash [5]Slot

	Fishing [3][3]Slot
	Horde   [3][5]Slot
}

type Slot struct {
	Pokemon  uint16
	MinLevel uint8
	MaxLevel uint8
}

func (s Slot) Species() int { return int(s.Pokemon & 0x7FF) }
func (s Slot) Form() int    { return int(s.Pokemon >> 11) }

func (s Slot) PokemonID() int {
	if s.Form() == 0 {
		return s.Species()
	}
	switch s.Species() {
	case 550:
		if s.Form() == 1 {
			return 10016 // basculin-blue-striped
		}
	case 664:
		return 664 // scatterbug
	case 669:
		return 669 // flabébé
	case 710: // pumpkaboo
		if 1 <= s.Form() && s.Form() <= 3 {
			return 10027 + s.Form() - 1
		}
	}
	panic(fmt.Sprintf("unknown form: pokemon %d, form %d", s.Species(), s.Form()))
}

func die(v ...interface{}) {
	fmt.Fprintln(os.Stderr, v...)
	os.Exit(1)
}

func main() {
	dburl := flag.String("import", "", "add encounters to `database`")
	flag.Parse()

	if flag.NArg() < 1 {
		die("usage: encounters [-import database] romfs/a/0/1/2")
	}

	f, err := os.Open(flag.Arg(0))
	if err != nil {
		die(err)
	}
	defer f.Close()

	files, err := garc.Files(f)
	if err != nil {
		die(err)
	}
	zonedata, err := ioutil.ReadAll(files[360])
	if err != nil {
		die(err)
	}

	var db *sql.DB
	var tx *sql.Tx
	if *dburl != "" {
		db, err = sql.Open("postgres", *dburl)
		if err != nil {
			die(err)
		}
		defer db.Close()

		tx, err = db.Begin()
		if err != nil {
			die(err)
		}
		defer tx.Rollback()

		/*
		exec := func(sql string, args ...interface{}) {
			_, err = tx.Exec(sql, args...)
			if err != nil {
				die(fmt.Sprintf("%v: %v", sql, err))
			}
		}
		exec(`DELETE FROM encounters WHERE version_id = $1`, VersionID)
		exec(`DELETE FROM location_area_prose WHERE location_area_id IN (SELECT la.id FROM location_areas la JOIN locations l ON la.location_id = l.id WHERE l.region_id=6)`)
		exec(`DELETE FROM location_areas WHERE id IN (SELECT la.id FROM location_areas la JOIN locations l on la.location_id = l.id WHERE l.region_id=6)`)
		exec(`SELECT setval('location_areas_id_seq', max(id)) FROM location_areas`)
		exec(`SELECT setval('encounters_id_seq', max(id)) FROM encounters`)
		*/
	}

	for i, g := range files {
		if i == 360 {
			continue
		}
		b, err := lz.Decode(g)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%d: %v\n", i, err)
			continue
		}
		off := int(le.Uint32(b[0x10:]))
		if off < 0 || off >= len(b) {
			continue
		}
		//fmt.Printf("% x\n", b[off:])
		var enc Encounter
		err = binary.Read(bytes.NewReader(b[off:]), le, &enc)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%d: %v\n", i, err)
			continue
		}
		if tx != nil {
			err = importEncounter(tx, i, &enc, zonedata[i*56:i*56+56])
			if err != nil {
				tx.Rollback()
				die(err)
			}
		} else {
			printEncounter(&enc, zonedata[i*56:i*56+56])
		}
	}
	if tx != nil {
		err = tx.Commit()
		if err != nil {
			die(err)
		}
	}
}

func importEncounter(tx *sql.Tx, index int, enc *Encounter, zd []byte) error {
	loc := int(zd[0x1C])
	areaID, err := addarea(tx, loc, index)
	if err != nil {
		return err
	}

	do := func(method string, slot []Slot, skip bool) {
		if err != nil {
			return
		}
		if skip {
			return
		}
		for i, t := range slot {
			err = addenc(tx, VersionID, areaID, method, i, t)
			if err != nil {
				return
			}
		}
	}

	do("walk", enc.Grass[:], enc.Header[0] == 0)
	do("yellow-flowers", enc.Flower[0][:], enc.Header[1] == 0)
	do("purple-flowers", enc.Flower[1][:], enc.Header[2] == 0)
	do("red-flowers", enc.Flower[2][:], enc.Header[3] == 0)
	do("rough-terrain", enc.Rough[:], enc.Header[4] == 0)
	do("surf", enc.Water[:], enc.Header[5] == 0)
	do("rock-smash", enc.RockSmash[:], enc.Header[6] == 0)
	do("old-rod", enc.Fishing[0][:], enc.Header[7] == 0)
	do("good-rod", enc.Fishing[1][:], enc.Header[8] == 0)
	do("super-rod", enc.Fishing[2][:], enc.Header[9] == 0)

	return err
}

func addarea(tx *sql.Tx, loc, index int) (areaID int, err error) {
	err = tx.QueryRow(`SELECT id FROM location_areas la JOIN location_game_indices lgi ON la.location_id = lgi.location_id WHERE la.identifier = $1 AND lgi.game_index = $2`,
		fmt.Sprintf("unknown-area-%d", index), loc).Scan(&areaID)
	if err == nil {
		return areaID, nil
	}
	if err != sql.ErrNoRows {
		return 0, err
	}

	err = tx.QueryRow(`INSERT INTO location_areas (identifier, location_id, game_index)
		SELECT $1, id, 0 FROM locations JOIN location_game_indices ON location_id = id AND generation_id = 6 WHERE game_index = $2
		RETURNING id`,
		fmt.Sprintf("unknown-area-%d", index), loc).Scan(&areaID)
	if err != nil {
		return 0, err
	}
	_, err = tx.Exec(`INSERT INTO location_area_prose (location_area_id, local_language_id, name) VALUES ($1, $2, $3)`,
		areaID, 9, fmt.Sprintf("Unknown Area %d", index))
	if err != nil {
		return areaID, err
	}
	return areaID, nil
}

func addenc(tx *sql.Tx, versionID, areaID int, method string, index int, slot Slot) error {
	_, err := tx.Exec(
		`INSERT INTO encounters
			(version_id, location_area_id, encounter_slot_id, pokemon_id, min_level, max_level)
			SELECT $1, $2, es.id, $3, $4, $5
			FROM encounter_slots es JOIN encounter_methods em ON es.encounter_method_id = em.id
			WHERE em.identifier = $6 AND es.version_group_id = $7 AND es.slot = $8`,
		versionID, areaID, slot.PokemonID(), slot.MinLevel, slot.MaxLevel,
		method, VersionGroupID, index)
	if err != nil {
		return fmt.Errorf("During query %v, %v, %v, %v, %v: %v", versionID, areaID, method, index, slot, err)
	}
	return nil
}

func printEncounter(enc *Encounter, zd []byte) {
	var b bytes.Buffer
	f := func(slots []Slot) string {
		b.Truncate(0)
		for i, t := range slots {
			if t.Pokemon == 0 {
				continue
			}
			if i != 0 {
				fmt.Fprint(&b, ", ")
			}
			if t.MinLevel == t.MaxLevel {
				fmt.Fprint(&b, t.MinLevel)
			} else {
				fmt.Fprint(&b, t.MinLevel, "-", t.MaxLevel)
			}
			fmt.Fprint(&b, " ", names.Species(t.Species()))
			if t.Form() != 0 {
				fmt.Fprintf(&b, " (form %d)", t.Form())
			}
		}
		return b.String()
	}
	fmt.Println(names.Location(int(zd[0x1C])))
	fmt.Printf("% x\n", enc.Header)
	fmt.Println("Grass:", f(enc.Grass[:]))
	fmt.Println("Yellow flowers:", f(enc.Flower[0][:]))
	fmt.Println("Purple flowers:", f(enc.Flower[1][:]))
	fmt.Println("Red flowers:", f(enc.Flower[2][:]))
	fmt.Println("Rough:", f(enc.Rough[:]))
	fmt.Println("Water:", f(enc.Water[:]))
	fmt.Println("Rock Smash:", f(enc.RockSmash[:]))
	fmt.Println("Old Rod:", f(enc.Fishing[0][:]))
	fmt.Println("Good Rod:", f(enc.Fishing[1][:]))
	fmt.Println("Super Rod:", f(enc.Fishing[2][:]))
	fmt.Println("Horde 1:", f(enc.Horde[0][:]))
	fmt.Println("Horde 2:", f(enc.Horde[1][:]))
	fmt.Println("Horde 3:", f(enc.Horde[2][:]))
	fmt.Println()
}
