package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"html/template"
	"os"
	"reflect"

	"xy/garc"
	"xy/names"
)

type Pokemon struct {
	Index int
	Name  string
	FormName string
	FullName string
	PokemonStats
}

type PokemonStats struct {
	HP             uint8
	Attack         uint8
	Defense        uint8
	Speed          uint8
	SpecialAttack  uint8
	SpecialDefense uint8
	Type           [2]uint8
	CatchRate      uint8
	ExpStage       uint8
	Effort         uint16
	Item           [3]uint16
	FemaleRate     uint8
	Hatch          uint8
	Friendship     uint8
	GrowthRate     uint8
	EggGroup       [2]uint8
	Ability        [3]uint8
	Unknown1B      uint8
	Form           uint16
	FormNameIndex  uint16
	FormCount      uint8
	Color          uint8
	Exp            uint16
	Height         uint16
	Weight         uint16
	TM             [16]uint8
	Tutor0         uint32
	Height2        uint16
	Unknown3E      uint16
	Extra          [16]uint8
}

func (p Pokemon) TypeText() string {
	if p.Type[0] == p.Type[1] {
		return names.Type(int(p.Type[0]))
	}
	return names.Type(int(p.Type[0])) + "/" + names.Type(int(p.Type[1]))
}

func (p Pokemon) EffortText() string {
	return fmt.Sprintf("%d/%d/%d/%d/%d/%d", p.Effort&3, p.Effort>>2&3, p.Effort>>4&3, p.Effort>>6&3, p.Effort>>8&3, p.Effort>>10&3)
}

func (p Pokemon) EggText() string {
	if p.EggGroup[0] == p.EggGroup[1] {
		return eggGroups[p.EggGroup[0]]
	}
	return eggGroups[p.EggGroup[0]] + "/" + eggGroups[p.EggGroup[1]]
}

func die(v ...interface{}) {
	fmt.Fprintln(os.Stderr, v...)
	os.Exit(1)
}

var t = template.Must(template.New("moves").Funcs(funcs).Parse(tmpltext))

func main() {
	filename := os.Args[1]
	f, err := os.Open(filename)
	if err != nil {
		die(err)
	}
	defer f.Close()

	files, err := garc.Files(f)
	if err != nil {
		die(err)
	}

	pokemon := make([]Pokemon, 0, len(files)-1)
	var p Pokemon
	for i, file := range files {
		if i == len(files)-1 {
			break
		}
		err := binary.Read(file, binary.LittleEndian, &p.PokemonStats)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
		p.Index = i
		p.Name = names.Species(i)
		if p.Index < len(formNames2) {
			p.FormName = formNames2[p.Index]
		}
		pokemon = append(pokemon, p)
	}

	for _, p := range pokemon {
		if p.Form != 0 {
			for j := 1; j < int(p.FormCount); j++ {
				pokemon[int(p.Form)+j-1].Name = p.Name
				pokemon[int(p.Form)+j-1].FormName =
					formNames[int(p.FormNameIndex)+j-1]
				pokemon[int(p.Form)+j-1].FullName =
					fullNames[int(p.FormNameIndex)+j-1]
			}
		}
	}

	err = t.Execute(os.Stdout, pokemon)
	if err != nil {
		die(err)
	}
}

var eggGroups = []string{
	"",
	"monster",
	"water1",
	"bug",
	"flying",
	"ground",
	"fairy",
	"plant",
	"humanshape",
	"water3",
	"mineral",
	"indeterminate",
	"water2",
	"ditto",
	"dragon",
	"no-eggs",
}

var formNames2 = [...]string{
	201: "One form",
	351: "Normal",
	382: "Kyogre",
	383: "Groudon",
	386: "Normal Forme",
	412: "Plant Cloak",
	413: "Plant Cloak",
	421: "Overcast Form",
	422: "West Sea",
	423: "West Sea",
	479: "Rotom",
	487: "Altered Forme",
	492: "Land Forme",
	493: "Arceus",
	550: "Red-Striped Form",
	555: "Standard Mode",
	585: "Spring Form",
	586: "Spring Form",
	641: "Incarnate Forme",
	642: "Incarnate Forme",
	645: "Incarnate Forme",
	646: "Kyurem",
	647: "Ordinary Form",
	648: "Aria Forme",
	649: "Genesect",
	666: "Icy Snow Pattern",
	669: "Red Flower",
	670: "Red Flower",
	671: "Red Flower",
	676: "Natural Form",
	678: "Male",
	681: "Shield Forme",
	710: "Average Size",
	711: "Average Size",
	716: "Neutral Mode",
	720: "Hoopa Confined",
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


var tmpltext = `<!DOCTYPE html>
<meta charset="utf-8">
<title>OR/AS pokémon base stats</title>
<style type="text/css">
  body { font-family: sans-serif; font-size: 16px; line-height: 1em; }
  table { border-collapse: collapse; white-space: nowrap; }
  thead { background: #B6C8E9; }
  tbody { border: 2px solid black; }
  tbody td, tbody th { border: 1px solid black; }
  td, th { padding: 0.3em; }
  td { text-align: right; }
  td.str { text-align: center; }
  td.list { text-align: left; }
  td.int { text-align: right; }
  td.hex { text-align: left; font-family: monospace; }
  tbody tr:hover { background: #DEE6F5; }
</style>

<table>
  <thead>
    <tr>
      <th>#</th>
      <th>Name</th>
      <th>Form Name</th>
      <th>Full Name</th>

      <th>HP</th>
      <th>Atk</th>
      <th>Def</th>
      <th>Spd</th>
      <th>SAtk</th>
      <th>SDef</th>
      <th>Type</th>
      <th>Catch</th>
      <th>Old</th>
      <th>Effort</th>
      <th>Item (50%)</th>
      <th>Item (5%)</th>
      <th>-</th>
      <th>♀</th>
      <th><img src="egg.png" alt="Egg"></th>
      <th>:3</th>
      <th>Egg Groups</th>
      <th>Growth</th>
      <th>Ability 0</th>
      <th>Ability 1</th>
      <th>Hidden Ability</th>
      <th>?</th>
      <th>Form</th>
      <th>Form</th>
      <th>#</th>
      <th>Color</th>
      <th>Exp.</th>
      <th>Height</th>
      <th>Weight</th>
      {{/*<th>TMs</th>*/}}
      {{/*<th>Tutors 0</th>*/}}
      <th>Height 2</th>
      <th>?</th>
      {{/*<th>Extra</th>*/}}

      <th>Name</th>
      <th>#</th>
    </tr>
  </thead>

  <tbody>
    {{range .}}
      <tr>
        <th>{{.Index}}</th>
        <th class=str>{{.Name}}</th>
        <td class=str>{{.FormName}}</th>
        <td class=str>{{.FullName}}</th>

        <td>{{.HP}}</td>
        <td>{{.Attack}}</td>
        <td>{{.Defense}}
        <td>{{.Speed}}</td>
        <td>{{.SpecialAttack}}</td>
        <td>{{.SpecialDefense}}</td>
        <td class=str>{{.TypeText}}</td>
        <td>{{.CatchRate}}</td>
        <td>{{.ExpStage}}</td>
        <td class=str>{{.EffortText}}</td>
        <td class=str>{{item (index .Item 0)}}</td>
        <td class=str>{{item (index .Item 1)}}</td>
        <td class=str>{{item (index .Item 2)}}</td>
        <td>{{.FemaleRate}}</td>
        <td>{{.Hatch}}</td>
        <td>{{.Friendship}}</td>
        <td class=str>{{.EggText}}</td>
        <td>{{.GrowthRate}}</td>
        <td class=str>{{index .Ability 0 | ability}}</td>
        <td class=str>{{if ne (index .Ability 0) (index .Ability 1)}}{{index .Ability 1 | ability}}{{end}}</td>
        <td class=str>{{index .Ability 2 | ability}}</td>
        <td>{{.Unknown1B}}</td>
        <td>{{.Form}}</td>
        <td>{{.FormNameIndex}}</td>
        <td>{{.FormCount}}</td>
        <td>{{.Color}}</td>
        <td>{{.Exp}}</td>
        <td>{{.Height}}</td>
        <td>{{.Weight}}</td>
        {{/*<td class=hex>{{bin .TM}}</td>*/}}
        {{/*<td class=hex>{{bin .Tutor0}}</td>*/}}
        <td>{{.Height2}}</td>
        <td>{{.Unknown3E}}</td>
        {{/*<td class=hex>{{printf "% x" .Extra}}</td>*/}}

        <th class=str>{{.Name}}</th>
        <th>{{.Index}}</th>
      </tr>
    {{end}}
  </tbody>
</table>

<script src="https://cdnjs.cloudflare.com/ajax/libs/jquery/2.1.4/jquery.js"></script>
<script src="https://cdnjs.cloudflare.com/ajax/libs/floatthead/1.2.10/jquery.floatThead.js"></script>
<script>$('table').floatThead();</script>
`

var (
	errBadComparisonType = errors.New("invalid type for comparison")
	errBadComparison     = errors.New("incompatible types for comparison")
	errNoComparison      = errors.New("missing argument for comparison")
)

var funcs = template.FuncMap{
	"eq": eq,
	"ne": ne,

	"item":    func(n uint16) string { return names.Item(int(n)) },
	"ability": func(n uint8) string { return names.Ability(int(n)) },
	"bin":     bin,
}

func bin(v interface{}) (string, error) {
	var b []byte
	switch v := v.(type) {
	case uint8:
		b = formatbin(b, uint64(v), 8)
	case uint16:
		b = formatbin(b, uint64(v), 16)
	case uint32:
		b = formatbin(b, uint64(v), 32)
	case uint64:
		b = formatbin(b, v, 64)
	default:
		if v := reflect.ValueOf(v); (v.Kind() == reflect.Slice || v.Kind() == reflect.Array) && v.Type().Elem().Kind() == reflect.Uint8 {
			n := v.Len()
			for i := 0; i < n; i++ {
				b = formatbin(b, v.Index(i).Uint(), 8)
			}
			return string(b), nil
		}
		return "", fmt.Errorf("bad type %T", v)
	}
	return string(b), nil
}

func formatbin(b []byte, v uint64, n int) []byte {
	for i := 0; i < n; i++ {
		if v&1 == 0 {
			b = append(b, '0')
		} else {
			b = append(b, '1')
		}
		v >>= 1
	}
	return b
}

type kind int

const (
	invalidKind kind = iota
	boolKind
	complexKind
	intKind
	floatKind
	integerKind
	stringKind
	uintKind
)

func basicKind(v reflect.Value) (kind, error) {
	switch v.Kind() {
	case reflect.Bool:
		return boolKind, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return intKind, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return uintKind, nil
	case reflect.Float32, reflect.Float64:
		return floatKind, nil
	case reflect.Complex64, reflect.Complex128:
		return complexKind, nil
	case reflect.String:
		return stringKind, nil
	}
	return invalidKind, errBadComparisonType
}

// eq evaluates the comparison a == b || a == c || ...
func eq(arg1 interface{}, arg2 ...interface{}) (bool, error) {
	v1 := reflect.ValueOf(arg1)
	k1, err := basicKind(v1)
	if err != nil {
		return false, err
	}
	if len(arg2) == 0 {
		return false, errNoComparison
	}
	for _, arg := range arg2 {
		v2 := reflect.ValueOf(arg)
		k2, err := basicKind(v2)
		if err != nil {
			return false, err
		}
		truth := false
		if k1 != k2 {
			// Special case: Can compare integer values regardless of type's sign.
			switch {
			case k1 == intKind && k2 == uintKind:
				truth = v1.Int() >= 0 && uint64(v1.Int()) == v2.Uint()
			case k1 == uintKind && k2 == intKind:
				truth = v2.Int() >= 0 && v1.Uint() == uint64(v2.Int())
			default:
				return false, errBadComparison
			}
		} else {
			switch k1 {
			case boolKind:
				truth = v1.Bool() == v2.Bool()
			case complexKind:
				truth = v1.Complex() == v2.Complex()
			case floatKind:
				truth = v1.Float() == v2.Float()
			case intKind:
				truth = v1.Int() == v2.Int()
			case stringKind:
				truth = v1.String() == v2.String()
			case uintKind:
				truth = v1.Uint() == v2.Uint()
			default:
				panic("invalid kind")
			}
		}
		if truth {
			return true, nil
		}
	}
	return false, nil
}

// ne evaluates the comparison a != b.
func ne(arg1, arg2 interface{}) (bool, error) {
	// != is the inverse of ==.
	equal, err := eq(arg1, arg2)
	return !equal, err
}
