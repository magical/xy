package main

import (
	"encoding/binary"
	"fmt"
	"html/template"
	"os"

	"xy/garc"
	"xy/names"
)

type Item struct {
	Index int
	Name  string
	ItemStats
}

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
func (m *Item) Flags() uint16        { return m.FlagsRaw >> 5 }

func die(v ...interface{}) {
	fmt.Fprintln(os.Stderr, v...)
	os.Exit(1)
}

var t = template.Must(template.New("items").Parse(tmpltext))

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
		//item.Name = names[i]
		items = append(items, item)
	}

	err = t.Execute(os.Stdout, items)
	if err != nil {
		die(err)
	}
}

var tmpltext = `<!DOCTYPE html>
<meta charset="utf-8">
<title>X/Y item struct</title>
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
  tr:hover { background: #DEE6F5; }
</style>

<table>
  <thead>
    <tr>
      <th>#</th>
      <th>Name</th>

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

        <td>{{.Price}}</td>
        <td>{{.Effect}}</td>
        <td>{{.EffectArg}}</td>
        <td>{{.NaturalGiftEffect}}</td>
        <td>{{.FlingEffect}}</td>
        <td>{{.FlingPower}}</td>
        <td>{{.NaturalGiftPower}}</td>
        <td>{{if ne .NaturalGiftType 31}}{{.NaturalGiftType}}{{end}}</td>
        <td class=hex>{{printf "%011b" .Flags}}</td>
        <td>{{.Unknown0A}}</td>
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
