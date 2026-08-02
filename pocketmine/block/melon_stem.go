package block

import "pocketmine-go/pocketmine/math"

// MelonStem is a port of pocketmine\block\MelonStem.
//
// AsItem should return VanillaItems.MELON_SEEDS() — needs the unported item package (see
// Block.GetDropsForCompatibleTool's doc comment), so it's left as Block's default for now.
type MelonStem struct {
	Stem
}

func NewMelonStem(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *MelonStem {
	m := &MelonStem{Stem{
		Crops: Crops{
			Flowable:     Flowable{Transparent{NewBlock(idInfo, name, typeInfo)}},
			AgeComponent: NewAgeComponent(CropsMaxAge),
		},
		Facing: math.Up,
	}}
	m.Init(m)
	return m
}

func (m *MelonStem) Clone() Behavior {
	c := *m
	c.rebind(&c)
	return &c
}

func (m *MelonStem) GetPlantTypeID() int { return MELON }

func (m *MelonStem) GetPlant() Behavior { return VanillaMelon() }
