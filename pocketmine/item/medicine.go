package item

import (
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/entity/effect"
)

// Medicine is a port of pocketmine\item\Medicine. GetResidue (VanillaItems.GLASS_BOTTLE()) and
// CanStartUsingItem (needs a real Player) aren't ported - see the Item interface's doc comment.
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

// OnConsume is a port of Medicine::onConsume: cures the medicine type's effect.
func (m *Medicine) OnConsume(consumer effect.Living) {
	consumer.GetEffects().Remove(m.MedicineTypeValue.GetCuredEffect())
}

func (m *Medicine) GetAdditionalEffects() []*effect.EffectInstance { return nil }
