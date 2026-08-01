package item

import blockutils "pocketmine-go/pocketmine/block/utils"

// BoatType is a port of pocketmine\item\BoatType. Its declaration order (Oak, Spruce, Birch,
// Jungle, Acacia, DarkOak, Mangrove) matches blockutils.WoodType's own first 7 values exactly, so
// GetWoodType is a direct conversion rather than a match table - same reasoning as
// blockutils.DyeColor's plain (non-inverted) ID mapping.
type BoatType int

const (
	BoatTypeOak BoatType = iota
	BoatTypeSpruce
	BoatTypeBirch
	BoatTypeJungle
	BoatTypeAcacia
	BoatTypeDarkOak
	BoatTypeMangrove
)

func (t BoatType) GetWoodType() blockutils.WoodType { return blockutils.WoodType(t) }

func (t BoatType) GetDisplayName() string { return t.GetWoodType().GetDisplayName() }
