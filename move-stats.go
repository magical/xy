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

type Move struct {
	Index int
	Name  string
	MoveStats
}

type MoveStats struct {
	Type            uint8
	Category        uint8
	DamageClassCode uint8
	Power           uint8
	Accuracy        uint8
	PP              uint8
	Priority        int8
	MultiHit        uint8
	StatusCode      int16
	StatusChance    uint8
	EffectLength    uint8
	EffectMinTurns  uint8
	EffectMaxTurns  uint8

	Crit       uint8
	Flinch     uint8
	Effect     uint16
	Recoil     int8
	Heal       int8
	Target     uint8
	StatType   [3]uint8
	StatStage  [3]int8
	StatChance [3]uint8
	Unknown1E  uint16
	Flags      uint32
}

func (m *Move) MultiHitMin() int { return int(m.MultiHit & 0xf) }
func (m *Move) MultiHitMax() int { return int(m.MultiHit >> 4) }
func (m *Move) IsMultiHit() bool { return m.MultiHit == 0 }

func (m *Move) TypeName() string { return names.Type(int(m.Type)) }

func (m *Move) DamageClass() string {
	switch m.DamageClassCode {
	case 0:
		return "status"
	case 1:
		return "physical"
	case 2:
		return "special"
	}
	return "unknown"
}

func (m *Move) Status() string {
	switch m.StatusCode {
	case 0:
		return ""
	case 1:
		return "parlyze"
	case 2:
		return "sleep"
	case 3:
		return "freeze"
	case 4:
		return "burn"
	case 5:
		return "poison"
	case 6:
		return "confuse"
	case 7:
		return "infatuate"
	case 8:
		return "trap"
	}
	return fmt.Sprint(m.StatusCode)
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

	file := files[0]
	var tmp [3]uint16
	err = binary.Read(file, binary.LittleEndian, &tmp)
	if err != nil {
		die(err)
	}
	n := int(tmp[1])
	var move Move
	moves := make([]Move, 0, n)
	file.Seek(int64(tmp[2]), 0)
	for i := 0; i < n; i++ {
		err := binary.Read(file, binary.LittleEndian, &move.MoveStats)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
		move.Index = i
		move.Name = names.Move(i)
		moves = append(moves, move)
	}

	type flag struct {
		Index int
		Name  string
		Has   []*Move
		Hasnt []*Move
	}
	flags := make([]flag, 16)
	flagnames := []string{
		0:  "contact",
		1:  "charge",
		2:  "recharge",
		3:  "protect",
		4:  "reflectable",
		5:  "snatch",
		6:  "mirror",
		7:  "punch",
		8:  "sound",
		9:  "gravity",
		10: "defrost",
		11: "distance",
		12: "heals",
		13: "substitute",
		14: "non-sky-battle",
		15: "",
	}
	for i := range flags {
		f := &flags[i]
		f.Index = i
		f.Name = flagnames[i]
		mask := uint32(1) << uint(i)
		for i, m := range moves {
			if m.Flags&mask != 0 {
				f.Has = append(f.Has, &moves[i])
			} else if i != 0 {
				f.Hasnt = append(f.Hasnt, &moves[i])
			}
		}
	}

	context := map[string]interface{}{
		"flags": flags,
		"moves": moves,
	}

	err = t.Execute(os.Stdout, context)
	if err != nil {
		die(err)
	}
}

var tmpltext = `<!DOCTYPE html>
<meta charset="utf-8">
<title>OR/AS moves</title>
<style type="text/css">
  body { font-family: sans-serif; font-size: 16px; line-height: 1em; }
  table { border-collapse: collapse; white-space: nowrap; }
  thead { background: #B6C8E9; }
  tbody { border: 2px solid black; }
  tbody td, tbody th { border: 1px solid black; }
  td, th { padding: 0.3em; }
  td { text-align: right; }
  td.str { text-align: left; }
  td.list { text-align: left; }
  td.int { text-align: right; }
  td.hex { font-family: monospace; }
  tbody tr:hover { background: #DEE6F5; }
</style>

<table>
  <thead>
    <tr>
      <th>#</th>
      <th>Name</th>

      <th>Type</th>
      <th>Category</th>
      <th>Damage Class</th>
      <th>Power</th>
      <th>Acc.</th>
      <th>PP</th>
      <th>Pri.</th>
      <th>Hits</th>
      <th>Status</th>
      <th>Status<br>Chance</th>
      <th>Status<br>Length</th>
      <th>Status<br>Turns</th>
      <th>Crit.</th>
      <th>Flinch</th>
      <th>Effect</th>
      <th>Recoil</th>
      <th>Heal</th>
      <th>Target</th>
      <th>Stat Type</th>
      <th>Stat Stage</th>
      <th>Stat Chance</th>
      <th>Unknown</th>
      <th>Flags</th>

      <th>Name</th>
      <th>#</th>
    </tr>
  </thead>

  <tbody>
    {{range .moves}}
      <tr>
        <th>{{.Index}}</th>
        <th class=str>{{.Name}}</th>

        <td class=str>{{.TypeName}}</td>
        <td>{{.Category}}</td>
        <td class=str>{{.DamageClass}}</td>
        <td>{{.Power}}</td>
        <td>{{.Accuracy}}</td>
        <td>{{.PP}}</td>
        <td>{{if ne .Priority 0}}{{.Priority}}{{end}}</td>
        <td>{{if not .IsMultiHit}}{{.MultiHitMin}}-{{.MultiHitMax}}{{end}}</td>
        <td class=str>{{.Status}}</td>
        <td>{{if ne .StatusChance 0}}{{.StatusChance}}%{{end}}</td>
        <td>{{if ne .EffectLength 0}}{{.EffectLength}}{{end}}</td>
        <td>{{if ne .EffectMinTurns 0}}{{.EffectMinTurns}}-{{.EffectMaxTurns}}{{end}}</td>
        <td>{{if ne .Crit 0 }}{{printf "%+d" .Crit}}{{end}}</td>
        <td>{{if ne .Flinch 0}}{{.Flinch}}%{{end}}</td>
        <td>{{.Effect}}</td>
        <td>{{if ne .Recoil 0}}{{.Recoil}}%{{end}}</td>
        <td>{{if ne .Heal 0}}{{.Heal}}%{{end}}</td>
        <td>{{.Target}}</td>
        <td class=list>{{.StatType}}</td>
        <td class=list>{{.StatStage}}</td>
        <td class=list>{{.StatChance}}</td>
        <td class=hex>{{printf "% x" .Unknown1E}}</td>
        <td class=hex>{{printf "%b" .Flags}}</td>

        <th class=str>{{.Name}}</th>
        <th>{{.Index}}</th>
      </tr>
    {{end}}
  </tbody>
</table>

{{range .flags}}
  <h1>Flag {{.Index}} {{.Name}} - {{len .Has}} moves</h1>
  {{if and (len .Has) (lt (len .Has) 400)}}
    <p>{{range $i, $_ := .Has}}{{if $i}}, {{end}}{{.Name}}{{end}}.</p>
  {{end}}
  {{if and (len .Hasnt) (lt (len .Hasnt) 400)}}
    <p>Every move <strong>except</strong>:
    <p>{{range $i, $_ := .Hasnt}}{{if $i}}, {{end}}{{.Name}}{{end}}.</p>
  {{end}}
{{end}}

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
