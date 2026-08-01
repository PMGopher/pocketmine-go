package block

import runtime "pocketmine-go/pocketmine/data/runtime"

// AnalogRedstoneSignalEmitter is a port of pocketmine\block\utils\AnalogRedstoneSignalEmitter.
type AnalogRedstoneSignalEmitter interface {
	GetOutputSignalStrength() int
	SetOutputSignalStrength(signalStrength int)
}

// AnalogRedstoneSignalEmitterComponent is a port of
// pocketmine\block\utils\AnalogRedstoneSignalEmitterTrait.
type AnalogRedstoneSignalEmitterComponent struct {
	SignalStrength int
}

func (a *AnalogRedstoneSignalEmitterComponent) DescribeSignalStrength(w runtime.DataDescriber) {
	w.BoundedIntAuto(0, 15, &a.SignalStrength)
}

func (a *AnalogRedstoneSignalEmitterComponent) GetOutputSignalStrength() int { return a.SignalStrength }

// SetOutputSignalStrength panics if out of range, mirroring the PHP original's
// \InvalidArgumentException (a programmer error at the call site).
func (a *AnalogRedstoneSignalEmitterComponent) SetOutputSignalStrength(signalStrength int) {
	if signalStrength < 0 || signalStrength > 15 {
		panic("Signal strength must be in range 0-15")
	}
	a.SignalStrength = signalStrength
}
