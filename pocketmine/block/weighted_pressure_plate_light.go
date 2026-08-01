package block

// WeightedPressurePlateLight is a port of pocketmine\block\WeightedPressurePlateLight.
//
// Deprecated in the PHP original too - retained for the same reason (a distinct registered block
// type, structurally identical to WeightedPressurePlate).
type WeightedPressurePlateLight struct {
	WeightedPressurePlate
}

func NewWeightedPressurePlateLight(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo, deactivationDelayTicks int, signalStrengthFactor float64) *WeightedPressurePlateLight {
	w := &WeightedPressurePlateLight{
		WeightedPressurePlate: WeightedPressurePlate{
			PressurePlate:        PressurePlate{Transparent: Transparent{NewBlock(idInfo, name, typeInfo)}, DeactivationDelayTicks: deactivationDelayTicks},
			SignalStrengthFactor: signalStrengthFactor,
		},
	}
	w.Init(w)
	return w
}

func (w *WeightedPressurePlateLight) Clone() Behavior {
	c := *w
	c.rebind(&c)
	return &c
}
