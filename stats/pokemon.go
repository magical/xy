package stats

// PokemonStats is the pokemon base stats structure found at
// a/2/1/8 in Pokémon X and Y, and
// a/1/9/5 in Pokémon Omega Ruby and Alpha Sapphire.
type PokemonStats struct {
	Stat       [6]uint8
	Type       [2]uint8
	CatchRate  uint8
	ExpStage   uint8
	RawEffort     uint16
	Item       [3]uint16
	FemaleRate uint8
	Hatch      uint8
	Friendship uint8
	GrowthRate uint8
	EggGroup   [2]uint8
	Ability    [3]uint8
	Unknown1B  uint8
	FormStats  uint16 // index of forms' base stats
	FormTotal  uint16 // cumulative FormCount
	FormCount  uint8
	Color      uint8
	Exp        uint16
	Height     uint16
	Weight     uint16
	TM         [16]uint8
	Tutor0     uint32
	Height2    uint16
	Unknown3E  uint16
	Extra      [16]uint8
}

func (p *PokemonStats) Effort() []int {
	e := p.RawEffort
	return []int{int(e&3), int(e>>2&3), int(e>>4&3), int(e>>6&3), int(e>>8&3), int(e>>10&3)}
}
