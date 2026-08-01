package item

import runtime "pocketmine-go/pocketmine/data/runtime"

// Medicine is a port of pocketmine\item\Medicine. OnConsume/GetResidue/CanStartUsingItem all need
// pieces that aren't ported (a real Living/Player with effect management, and the item registry
// for GetResidue's VanillaItems.GLASS_BOTTLE()) - see the Item interface's doc comment.
type Medicine struct {
	ItemBase

	MedicineTypeValue MedicineType
}

func NewMedicine(identifier ItemIdentifier, name string) *Medicine {
	m := &Medicine{MedicineTypeValue: MedicineTypeEyeDrops}
	m.Init(m, identifier, name)
	return m
}

func (m *Medicine) Clone() Item {
	c := *m
	c.rebind(&c)
	return &c
}

func (m *Medicine) GetType() MedicineType { return m.MedicineTypeValue }

func (m *Medicine) SetType(t MedicineType) { m.MedicineTypeValue = t }

func (m *Medicine) GetMaxStackSize() int { return 1 }

func (m *Medicine) describeState(w runtime.DataDescriber) {
	t := int(m.MedicineTypeValue)
	w.BoundedIntAuto(int(MedicineTypeAntidote), int(MedicineTypeTonic), &t)
	m.MedicineTypeValue = MedicineType(t)
}
