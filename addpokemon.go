package main

import (
	"database/sql"
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/juju/errors"
	_ "github.com/lib/pq"

	"xy/stats"
	"xy/util"
)

var (
	gen     int
	pair    int
	version int
)

func main() {
	if len(os.Args) < 5 {
		fmt.Fprintln(os.Stderr, "usage: addpokemon database version path/to/garc n")
		os.Exit(1)
	}
	dbname := os.Args[1]
	versionname := os.Args[2]
	garcname := os.Args[3]
	numbers := os.Args[4:]

	g, err := util.OpenGARC(garcname)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer g.Close()

	db, err := sql.Open("postgres", fmt.Sprintf("postgres:///%s?sslmode=disable", dbname))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer db.Close()

	err = db.QueryRow(
		`SELECT v.id, vg.id, g.id
		FROM versions v
		JOIN version_groups vg on v.version_group_id = vg.id
		JOIN generations g on vg.generation_id = g.id
		WHERE v.identifier = $1`,
		versionname).Scan(&version, &pair, &gen)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println(gen, pair, version)

	pokemon, err := readPokemon(g)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	tx, err := db.Begin()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer tx.Rollback()
	for i := range numbers {
		n, err := strconv.ParseInt(numbers[i], 0, 0)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		p := pokemon[n]
		fmt.Println(i, p.Name)
		if err := AddForm(tx, p); err != nil {
			fmt.Fprintln(os.Stderr, err)
			break
		}
	}
	if err := tx.Commit(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type Pokemon struct {
	Index   int
	Species int
	Form    int

	Name string
	FormName string

	stats.PokemonStats
}

func readPokemon(g *util.GARC) ([]*Pokemon, error) {
	list := make([]*Pokemon, 0, len(g.Files)-1)
	sz := int64(binary.Size(stats.PokemonStats{}))
	for i, f := range g.Files {
		if i == len(g.Files)-1 {
			break
		}

		var p Pokemon
		if f.Size() != sz {
			return nil, fmt.Errorf("size mismatch")
		}
		err := binary.Read(f, binary.LittleEndian, &p.PokemonStats)
		if err != nil {
			return nil, err
		}
		list = append(list, &p)
	}
	for i, p := range list {
		p.Index = i
		if p.FormStats != 0 && p.FormCount > 1 {
			forms := list[p.FormStats:]
			for j := 1; j < int(p.FormCount); j++ {
				forms[j-1].Species = i
				forms[j-1].Form = j
				forms[j-1].Name = fullNames[int(p.FormTotal)+j-1]
				forms[j-1].FormName = formNames[int(p.FormTotal)+j-1]
			}
		}
	}
	return list, nil
}

func AddForm(tx *sql.Tx, p *Pokemon) error {
	var id int     // pokemon id
	var formid int // pokemon_form id
	if err := tx.QueryRow(`select max(id)+1 from pokemon`).Scan(&id); err != nil {
		return err
	}
	if err := tx.QueryRow(`select max(id)+1 from pokemon_forms`).Scan(&formid); err != nil {
		return err
	}

	// Since we are only adding alternate forms,
	// the pokemon identifier is the same as the pokemon_form identifier
	ident := nameToIdent(p.Name)

	_, err := tx.Exec(`INSERT INTO pokemon
		(id, identifier, species_id, height, weight, base_experience, "order", is_default)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		id, ident, p.Species, p.Height, p.Weight, p.Exp, 0, p.Species == p.Index)
	if err != nil {
		return errors.Annotate(err, "adding pokemon")
	}

	_, err = tx.Exec(`INSERT INTO pokemon_forms
		(id, identifier, form_identifier, pokemon_id, introduced_in_version_group_id, is_default, is_battle_only, is_mega, form_order, "order")
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		formid, ident, ident, id, pair, true, false, false, 0, 0)
	if err != nil {
		return errors.Annotate(err, "adding pokemon_form")
	}

	err = addTypes(tx, id, p.Type[:])
	if err != nil {
		return errors.Annotate(err, "adding types")
	}

	err = addAbilities(tx, id, p.Ability[:])
	if err != nil {
		return errors.Annotate(err, "adding abilities")
	}

	/*
	if p.Item[0] == p.Item[1] {
		err = addItems(tx, id, p.Item[:1], 100)
	} else {
		err = addItems(tx, id, p.Item[:2], 50, 5)
	}
	if err != nil {
		return errors.Annotate(err, "adding items")
	}
	*/

	err = addStats(tx, id, p.Stat[:], p.Effort())
	if err != nil {
		return errors.Annotate(err, "adding stats")
	}

	// Egg groups are per-species
	// Colors are per-species
	// Names are i18n

	err = addPokemonGameIndex(tx, id, p.Index)
	if err != nil {
		return err
	}

	err = addPokemonFormIndex(tx, formid, p.Form)
	if err != nil {
		return err
	}
	err = addNames(tx, formid, p.Name, p.FormName)
	if err != nil {
		return err
	}

	return nil
}

func addTypes(tx *sql.Tx, pokemon int, types []uint8) error {
	const sql = `INSERT INTO pokemon_types (pokemon_id, slot, type_id)
		VALUES ($1, $3,
			(select type_id from type_game_indices
			where game_index = $2 and generation_id = $4))`

	for slot, id := range types {
		if slot > 0 && id == types[0] {
			continue
		}
		_, err := tx.Exec(sql, pokemon, id, slot+1, gen)
		if err != nil {
			return err
		}
	}
	return nil
}

func addAbilities(tx *sql.Tx, pokemon int, abilities []uint8) error {
	const sql = `INSERT INTO pokemon_abilities
		(pokemon_id, ability_id, slot, is_hidden)
		VALUES ($1, $2, $3, $4)`

	for slot, id := range abilities {
		if slot > 0 && id == abilities[0] {
			continue
		}
		_, err := tx.Exec(sql, pokemon, id, slot+1, slot == 2)
		if err != nil {
			return err
		}
	}
	return nil
}

func addItems(tx *sql.Tx, pokemon int, items []uint16, rates ...int) error {
	const sql = `INSERT INTO pokemon_items (pokemon_id, version_id, rarity, item_id)
		VALUES ($1, $5, $3,
			(select item_id from item_game_indices
			where game_index = $2 and generation_id = $4))`

	for slot, id := range items {
		if id == 0 {
			continue
		}
		_, err := tx.Exec(sql, pokemon, id, rates[slot], gen, version)
		if err != nil {
			return err
		}
	}
	return nil
}

func addStats(tx *sql.Tx, pokemon int, stats []uint8, effort []int) error {
	const sql = `INSERT INTO pokemon_stats (pokemon_id, stat_id, base_stat, effort)
		VALUES ($1, (select id from stats where game_index = $2+1), $3, $4)`
	for i := range stats {
		_, err := tx.Exec(sql, pokemon, i, stats[i], effort[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func addPokemonGameIndex(tx *sql.Tx, pokemon, index int) error {
	_, err := tx.Exec(
		`INSERT INTO pokemon_game_indices
		(pokemon_id, game_index, version_id)
		VALUES ($1, $2, $3)`,
		pokemon, index, version)
	return err
}

func addPokemonFormIndex(tx *sql.Tx, form, index int) error {
	_, err := tx.Exec(
		`INSERT INTO pokemon_form_generations
		(pokemon_form_id, game_index, generation_id)
		VALUES ($1, $2, $3)`,
		form, index, gen)
	return err
}

func addNames(tx *sql.Tx, form int, name, formname string) error {
	_, err := tx.Exec(
		`INSERT INTO pokemon_form_names
		(pokemon_form_id, local_language_id, form_name, pokemon_name)
		VALUES ($1, $2, $3, $4)`,
		form, 9, formname, name)
	return err
}

func nameToIdent(s string) string {
	replace := strings.NewReplacer(
		"é", "e",
		" ", "-",
		".", "",
		",", "",
	)
	s = strings.ToLower(s)
	s = replace.Replace(s)
	return s
}

var formNames = []string{
	"Mega Venusaur",
	"Mega Charizard X",
	"Mega Charizard Y",
	"Mega Blastoise",
	"Mega Beedrill",
	"Mega Pidgeot",
	"Pikachu Rock Star",
	"Pikachu Belle",
	"Pikachu Pop Star",
	"Pikachu, Ph.D.",
	"Pikachu Libre",
	"Cosplay Pikachu",
	"Mega Alakazam",
	"Mega Slowbro",
	"Mega Gengar",
	"Mega Kangaskhan",
	"Mega Pinsir",
	"Mega Gyarados",
	"Mega Aerodactyl",
	"Mega Mewtwo X",
	"Mega Mewtwo Y",
	"Mega Ampharos",
	"One form",
	"One form",
	"One form",
	"One form",
	"One form",
	"One form",
	"One form",
	"One form",
	"One form",
	"One form",
	"One form",
	"One form",
	"One form",
	"One form",
	"One form",
	"One form",
	"One form",
	"One form",
	"One form",
	"One form",
	"One form",
	"One form",
	"One form",
	"One form",
	"One form",
	"One form",
	"One form",
	"Mega Steelix",
	"Mega Scizor",
	"Mega Heracross",
	"Mega Houndoom",
	"Mega Tyranitar",
	"Mega Sceptile",
	"Mega Blaziken",
	"Mega Swampert",
	"Mega Gardevoir",
	"Mega Sableye",
	"Mega Mawile",
	"Mega Aggron",
	"Mega Medicham",
	"Mega Manectric",
	"Mega Sharpedo",
	"Mega Camerupt",
	"Mega Altaria",
	"Sunny Form",
	"Rainy Form",
	"Snowy Form",
	"Mega Banette",
	"Mega Absol",
	"Mega Glalie",
	"Mega Salamence",
	"Mega Metagross",
	"Mega Latias",
	"Mega Latios",
	"Primal Reversion",
	"Primal Reversion",
	"Mega Rayquaza",
	"Attack Forme",
	"Defense Forme",
	"Speed Forme",
	"Sandy Cloak",
	"Trash Cloak",
	"Sandy Cloak",
	"Trash Cloak",
	"Sunshine Form",
	"East Sea",
	"East Sea",
	"Mega Lopunny",
	"Mega Garchomp",
	"Mega Lucario",
	"Mega Abomasnow",
	"Mega Gallade",
	"Heat Rotom",
	"Wash Rotom",
	"Frost Rotom",
	"Fan Rotom",
	"Mow Rotom",
	"Origin Forme",
	"Sky Forme",
	// Arceus
	"Mega Audino",
	"Blue-Striped Form",
	"Zen Mode",
	"Summer Form",
	"Autumn Form",
	"Winter Form",
	"Summer Form",
	"Autumn Form",
	"Winter Form",
	"Therian Forme",
	"Therian Forme",
	"Therian Forme",
	"White Kyurem",
	"Black Kyurem",
	"Resolute Form",
	"Pirouette Forme",
	"Genesect",
	"Genesect",
	"Genesect",
	"Genesect",
	"Polar Pattern",
	"Tundra Pattern",
	"Continental Pattern",
	"Garden Pattern",
	"Elegant Pattern",
	"Meadow Pattern",
	"Modern Pattern",
	"Marine Pattern",
	"Archipelago Pattern",
	"High Plains Pattern",
	"Sandstorm Pattern",
	"River Pattern",
	"Monsoon Pattern",
	"Savanna Pattern",
	"Sun Pattern",
	"Ocean Pattern",
	"Jungle Pattern",
	"Fancy Pattern",
	"Poké Ball Pattern",
	"Yellow Flower",
	"Orange Flower",
	"Blue Flower",
	"White Flower",
	"Yellow Flower",
	"Orange Flower",
	"Blue Flower",
	"White Flower",
	"Eternal Flower",
	"Yellow Flower",
	"Orange Flower",
	"Blue Flower",
	"White Flower",
	"Heart Trim",
	"Star Trim",
	"Diamond Trim",
	"Debutante Trim",
	"Matron Trim",
	"Dandy Trim",
	"La Reine Trim",
	"Kabuki Trim",
	"Pharaoh Trim",
	"Female",
	"Blade Forme",
	"Small Size",
	"Large Size",
	"Super Size",
	"Small Size",
	"Large Size",
	"Super Size",
	"Active Mode",
	"Mega Diancie",
	"Hoopa Unbound",
}

var fullNames = []string{
	"Mega Venusaur",
	"Mega Charizard X",
	"Mega Charizard Y",
	"Mega Blastoise",
	"Mega Beedrill",
	"Mega Pidgeot",
	"Pikachu Rock Star",
	"Pikachu Belle",
	"Pikachu Pop Star",
	"Pikachu, Ph.D.",
	"Pikachu Libre",
	"Cosplay Pikachu",
	"Mega Alakazam",
	"Mega Slowbro",
	"Mega Gengar",
	"Mega Kangaskhan",
	"Mega Pinsir",
	"Mega Gyarados",
	"Mega Aerodactyl",
	"Mega Mewtwo X",
	"Mega Mewtwo Y",
	"Mega Ampharos",
	"Unown",
	"Unown",
	"Unown",
	"Unown",
	"Unown",
	"Unown",
	"Unown",
	"Unown",
	"Unown",
	"Unown",
	"Unown",
	"Unown",
	"Unown",
	"Unown",
	"Unown",
	"Unown",
	"Unown",
	"Unown",
	"Unown",
	"Unown",
	"Unown",
	"Unown",
	"Unown",
	"Unown",
	"Unown",
	"Unown",
	"Unown",
	"Mega Steelix",
	"Mega Scizor",
	"Mega Heracross",
	"Mega Houndoom",
	"Mega Tyranitar",
	"Mega Sceptile",
	"Mega Blaziken",
	"Mega Swampert",
	"Mega Gardevoir",
	"Mega Sableye",
	"Mega Mawile",
	"Mega Aggron",
	"Mega Medicham",
	"Mega Manectric",
	"Mega Sharpedo",
	"Mega Camerupt",
	"Mega Altaria",
	"Sunny Castform",
	"Rainy Castform",
	"Snowy Castform",
	"Mega Banette",
	"Mega Absol",
	"Mega Glalie",
	"Mega Salamence",
	"Mega Metagross",
	"Mega Latias",
	"Mega Latios",
	"Primal Kyogre",
	"Primal Groudon",
	"Mega Rayquaza",
	"Attack Deoxys",
	"Defense Deoxys",
	"Speed Deoxys",
	"Sandy Burmy",
	"Trash Burmy",
	"Sandy Wormadam",
	"Trash Wormadam",
	"Sunshine Cherrim",
	"East Shellos",
	"East Gastrodon",
	"Mega Lopunny",
	"Mega Garchomp",
	"Mega Lucario",
	"Mega Abomasnow",
	"Mega Gallade",
	"Heat Rotom",
	"Wash Rotom",
	"Frost Rotom",
	"Fan Rotom",
	"Mow Rotom",
	"Origin Giratina",
	"Sky Shaymin",
	// Arceus
	"Mega Audino",
	"Blue-Striped Form",
	"Zen Mode",
	"Summer Deerling",
	"Autumn Deerling",
	"Winter Deerling",
	"Summer Sawsbuck",
	"Autumn Sawsbuck",
	"Winter Sawsbuck",
	"Therian Forme",
	"Therian Forme",
	"Therian Forme",
	"White Kyurem",
	"Black Kyurem",
	"Resolute Form",
	"Pirouette Forme",
	"Genesect",
	"Genesect",
	"Genesect",
	"Genesect",
	"Polar Pattern",
	"Tundra Pattern",
	"Continental Pattern",
	"Garden Pattern",
	"Elegant Pattern",
	"Meadow Pattern",
	"Modern Pattern",
	"Marine Pattern",
	"Archipelago Pattern",
	"High Plains Pattern",
	"Sandstorm Pattern",
	"River Pattern",
	"Monsoon Pattern",
	"Savanna Pattern",
	"Sun Pattern",
	"Ocean Pattern",
	"Jungle Pattern",
	"Fancy Pattern",
	"Poké Ball Pattern",
	"Yellow Flabébé",
	"Orange Flabébé",
	"Blue Flabébé",
	"White Flabébé",
	"Yellow Floette",
	"Orange Floette",
	"Blue Floette",
	"White Floette",
	"Eternal Floette",
	"Yellow Florges",
	"Orange Florges",
	"Blue Florges",
	"White Florges",
	"Heart Trim",
	"Star Trim",
	"Diamond Trim",
	"Debutante Trim",
	"Matron Trim",
	"Dandy Trim",
	"La Reine Trim",
	"Kabuki Trim",
	"Pharaoh Trim",
	"Female",
	"Blade Forme",
	"Small Size",
	"Large Size",
	"Super Size",
	"Small Size",
	"Large Size",
	"Super Size",
	"Active Xernias",
	"Mega Diancie",
	"Hoopa Unbound",
}
