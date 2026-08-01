package block

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
)

// Froglight is a port of pocketmine\block\Froglight.
type Froglight struct {
	SimplePillar

	FroglightTypeValue blockutils.FroglightType
}

func NewFroglight(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Froglight {
	f := &Froglight{
		SimplePillar: SimplePillar{
			Opaque:                  Opaque{NewBlock(idInfo, name, typeInfo)},
			PillarRotationComponent: NewPillarRotationComponent(),
		},
		FroglightTypeValue: blockutils.FroglightTypeOchre,
	}
	f.Init(f)
	return f
}

func (f *Froglight) Clone() Behavior {
	c := *f
	c.rebind(&c)
	return &c
}

func (f *Froglight) DescribeBlockItemState(w runtime.DataDescriber) {
	t := int(f.FroglightTypeValue)
	w.BoundedIntAuto(int(blockutils.FroglightTypeOchre), int(blockutils.FroglightTypeVerdant), &t)
	f.FroglightTypeValue = blockutils.FroglightType(t)
}

func (f *Froglight) GetFroglightType() blockutils.FroglightType { return f.FroglightTypeValue }

func (f *Froglight) SetFroglightType(froglightType blockutils.FroglightType) {
	f.FroglightTypeValue = froglightType
}

func (f *Froglight) GetLightLevel() int { return 15 }
