package names

func Type(n int) string {
	if 0 <= n && n < len(typeNames) {
		return typeNames[n]
	}
	return ""
}

var typeNames = []string{
	"Normal",
	"Fighting",
	"Flying",
	"Poison",
	"Ground",
	"Rock",
	"Bug",
	"Ghost",
	"Steel",
	"Fire",
	"Water",
	"Grass",
	"Electric",
	"Psychic",
	"Ice",
	"Dragon",
	"Dark",
	"Fairy",
}
