package item

// MedicineType is a port of pocketmine\item\MedicineType. GetCuredEffect isn't ported - Effect
// (entity/effect package) isn't ported, same gap documented throughout this port wherever an
// Effect/EffectInstance would be constructed. Only GetDisplayName (plain string data) is ported.
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
