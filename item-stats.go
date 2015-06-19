package main

import (
	"encoding/binary"
	"fmt"
	"html/template"
	"os"
	"strconv"
	"strings"

	"xy/garc"
	"xy/names"
)

type Item struct {
	Index int
	Name  string
	Icon  int
	ItemStats
}

// ItemStats is the item stat structure found at
// a/2/2/0 in Pokémon X and Y, and
// a/1/9/7 in Pokémon Omega Ruby and Alpha Sapphire.
type ItemStats struct {
	PriceRaw          uint16
	Effect            uint8
	EffectArg         uint8
	NaturalGiftEffect uint8
	FlingEffect       uint8
	FlingPower        uint8
	NaturalGiftPower  uint8
	FlagsRaw          uint16
	Unknown0A         uint8
	Unknown0B         uint8
	Unknown0C         uint8
	Unknown0D         uint8
	Unknown0E         uint8
	Order             uint8

	Status1   uint32
	Status2   uint16
	Status3   uint8

	Effort     [6]int8
	HP         uint8
	PP         uint8
	Friendship [3]int8
}

func (m *Item) Price() uint          { return uint(m.PriceRaw) * 10 }
func (m *Item) Status() uint64       { return uint64(m.Status1) | uint64(m.Status2)<<32 | uint64(m.Status3)<<48 }
func (m *Item) NaturalGiftType() int { return int(m.FlagsRaw & 31) }
func (m *Item) NaturalGiftTypeName() string { return names.Type(int(m.FlagsRaw & 31)) }
func (m *Item) Flags() uint16        { return m.FlagsRaw >> 5 }
// Flags:
// 	2 
// 	3 berry / tm
// 	4 key item
//	5 nothing
// 	6 ball
// 	7 battle item
// 	8 restores HP or PP
// 	9 restores status

func die(v ...interface{}) {
	fmt.Fprintln(os.Stderr, v...)
	os.Exit(1)
}

var t = template.Must(template.New("items").Funcs(funcs).Parse(tmpltext))

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

	var iconmap []uint32
	if len(os.Args) > 2 {
		iconmap, err = readiconmap(os.Args[2], len(files))
		if err != nil {
			die(err)
		}
	}

	var item Item
	items := make([]Item, 0, len(files))
	for i, file := range files {
		err := binary.Read(file, binary.LittleEndian, &item.ItemStats)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}
		item.Index = i
		item.Name = names.Item(i)
		if iconmap != nil {
			item.Icon = int(iconmap[i])
		} else {
			item.Icon = -1
		}
		//item.Name = names[i]
		items = append(items, item)
	}

	err = t.Execute(os.Stdout, items)
	if err != nil {
		die(err)
	}
}

func readiconmap(filename string, n int) ([]uint32, error) {
	filename, offstr := partition(filename, ":")

	off, err := strconv.ParseInt(offstr, 0, 64)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if _, err := f.Seek(off, 0); err != nil {
		return nil, err
	}

	m := make([]uint32, n)
	err = binary.Read(f, binary.LittleEndian, m)
	return m, err
}

func partition(s string, sep string) (front, back string) {
	i := strings.Index(s, sep)
	if i >= 0 {
		return s[:i], s[i+1:]
	}
	return s, ""
}

func flags(v interface{}, s string) string {
	switch v := v.(type) {
	case uint16:
		return formatFlags(uint32(v), s)
	}
	return "error"
}

func formatFlags(u uint32, s string) string {
	var b [64]byte
	for i := 0; i < len(s); i++ {
		if u & 1 == 0 {
			b[i] = '-'
		} else {
			b[i] = s[i]
		}
		u = u >> 1
	}
	return string(b[:len(s)])
}

var funcs = template.FuncMap{
	"flags": flags,
}

var tmpltext = `<!DOCTYPE html>
<meta charset="utf-8">
<title>OR/AS item struct</title>
<style type="text/css">
  body { font-family: sans-serif; font-size: 16px; line-height: 1em; }
  table { border-collapse: collapse; white-space: nowrap; }
  tbody { border: 2px solid black; }
  tbody td, tbody th { border: 1px solid black; }
  td, th { padding: 0.3em; }
  td { text-align: right; }
  td.str { text-align: left; }
  td.list { text-align: left; }
  td.int { text-align: right; }
  td.hex { font-family: monospace; }
  td.icon { padding: 0; text-align: center; }
  td img { vertical-align: middle; }
  tr:hover { background: #DEE6F5; }
</style>

<table>
  <thead>
    <tr>
      <th>#</th>
      <th>Name</th>
      <th>Icon</th>

      <th>Price</th>
      <th>Effect</th>
      <th>Arg</th>
      <th>Ntl.Gift<br>effect</th>
      <th>Fling<br>effect</th>
      <th>Fling<br>power</th>
      <th>Ntl.Gift<br>power</th>
      <th>Ntl.Gift<br>type</th>
      <th>Flags</th>
      <th>0A</th>
      <th>0B</th>
      <th>0C</th>
      <th>0D</th>
      <th>0E</th>
      <th>Order</th>
      <th>Status</th>
      <th>Effort</th>
      <th>HP</th>
      <th>PP</th>
      <th>Friendship</th>

      <th>Name</th>
      <th>#</th>
    </tr>
  </thead>

  <tbody>
    {{range .}}
      <tr>
        <th>{{.Index}}</th>
        <th class=str>{{.Name}}</th>
        <td class=icon>{{if ne .Icon -1}}<img src="items/{{.Icon}}.png">{{end}}</td>

        <td>{{.Price}}</td>
        <td>{{.Effect}}</td>
        <td>{{.EffectArg}}</td>
        <td>{{.NaturalGiftEffect}}</td>
        <td>{{.FlingEffect}}</td>
        <td>{{.FlingPower}}</td>
        <td>{{.NaturalGiftPower}}</td>
        <td class=str>{{if ne .NaturalGiftType 31}}{{.NaturalGiftTypeName}}{{end}}</td>
        <td class=hex>{{flags .Flags "012mk%bths%"}}</td>
        <td class=hex>{{printf "%x" .Unknown0A}}</td>
        <td>{{.Unknown0B}}</td>
        <td>{{.Unknown0C}}</td>
        <td>{{.Unknown0D}}</td>
        <td>{{.Unknown0E}}</td>
        <td>{{.Order}}</td>
        <td class=hex>{{printf "%014x" .Status}}</td>
        <td class=list>{{.Effort}}</td>
        <td>{{.HP}}</td>
        <td>{{.PP}}</td>
        <td class=list>{{.Friendship}}</td>

        <th class=str>{{.Name}}</th>
        <th>{{.Index}}</th>
      </tr>
    {{end}}
  </tbody>
</table>

`
