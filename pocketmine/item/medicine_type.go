package item

import (
	"pocketmine-go/pocketmine/entity/effect"
)

// MedicineType is a port of pocketmine\item\MedicineType.
type MedicineType int

const (
	MedicineTypeAntidote MedicineType = iota
	MedicineTypeElixir
	MedicineTypeEyeDrops
	MedicineTypeTonic
)

var medicineTypeDisplayNames = map[MedicineType]string{
	MedicineTypeAntidote: "Antidote",
	MedicineTypeElixir:   "Elixir",
	MedicineTypeEyeDrops: "Eye Drops",
	MedicineTypeTonic:    "Tonic",
}

func (t MedicineType) GetDisplayName() string { return medicineTypeDisplayNames[t] }

// GetCuredEffect is a port of MedicineType::getCuredEffect.
func (t MedicineType) GetCuredEffect() effect.Effect {
	switch t {
	case MedicineTypeAntidote:
		return effect.VanillaPoison()
	case MedicineTypeElixir:
		return effect.VanillaWeakness()
	case MedicineTypeEyeDrops:
		return effect.VanillaBlindness()
	default:
		return effect.VanillaNausea()
	}
}
