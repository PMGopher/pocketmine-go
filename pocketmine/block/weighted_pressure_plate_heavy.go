package block

// WeightedPressurePlateHeavy is a port of pocketmine\block\WeightedPressurePlateHeavy.
//
// Deprecated in the PHP original too - retained for the same reason (a distinct registered block
// type, structurally identical to WeightedPressurePlate).
type WeightedPressurePlateHeavy struct {
	WeightedPressurePlate
}

func NewWeightedPressurePlateHeavy(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo, deactivationDelayTicks int, signalStrengthFactor float64) *WeightedPressurePlateHeavy {
	w := &WeightedPressurePlateHeavy{
		WeightedPressurePlate: WeightedPressurePlate{
			PressurePlate:        PressurePlate{Transparent: Transparent{NewBlock(idInfo, name, typeInfo)}, DeactivationDelayTicks: deactivationDelayTicks},
			SignalStrengthFactor: signalStrengthFactor,
		},
	}
	w.Init(w)
	return w
}

func (w *WeightedPressurePlateHeavy) Clone() Behavior {
	c := *w
	c.rebind(&c)
	return &c
}
