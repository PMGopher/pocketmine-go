package block

import (
	stdmath "math"

	runtime "pocketmine-go/pocketmine/data/runtime"
)

// WeightedPressurePlate is a port of pocketmine\block\WeightedPressurePlate.
type WeightedPressurePlate struct {
	PressurePlate
	AnalogRedstoneSignalEmitterComponent

	// SignalStrengthFactor: number of entities on the plate is divided by this value to get
	// signal strength.
	SignalStrengthFactor float64
}

func NewWeightedPressurePlate(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo, deactivationDelayTicks int, signalStrengthFactor float64) *WeightedPressurePlate {
	w := &WeightedPressurePlate{
		PressurePlate:        PressurePlate{Transparent: Transparent{NewBlock(idInfo, name, typeInfo)}, DeactivationDelayTicks: deactivationDelayTicks},
		SignalStrengthFactor: signalStrengthFactor,
	}
	w.Init(w)
	return w
}

func (w *WeightedPressurePlate) Clone() Behavior {
	c := *w
	c.rebind(&c)
	return &c
}

func (w *WeightedPressurePlate) DescribeBlockOnlyState(d runtime.DataDescriber) {
	w.DescribeSignalStrength(d)
}

func (w *WeightedPressurePlate) hasOutputSignal() bool { return w.SignalStrength > 0 }

func (w *WeightedPressurePlate) calculatePlateState(entities []Entity) (Behavior, *bool) {
	newSignalStrength := int(stdmath.Ceil(float64(len(entities)) * w.SignalStrengthFactor))
	if newSignalStrength > 15 {
		newSignalStrength = 15
	}
	if newSignalStrength < 0 {
		newSignalStrength = 0
	}
	if newSignalStrength == w.SignalStrength {
		return w.self, nil
	}

	wasActive := w.SignalStrength != 0
	isActive := newSignalStrength != 0

	clone := w.self.Clone()
	clone.(AnalogRedstoneSignalEmitter).SetOutputSignalStrength(newSignalStrength)

	if wasActive != isActive {
		return clone, &isActive
	}
	return clone, nil
}
