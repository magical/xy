package stats

// MoveStats is the move stat structure found at
// a/2/1/2 in Pokémon X and Y, and
// a/1/8/9 in Pokémon Omega Ruby and Alpha Sapphire.
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
