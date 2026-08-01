package block

import runtime "pocketmine-go/pocketmine/data/runtime"

// SimplePressurePlate is a port of pocketmine\block\SimplePressurePlate.
//
// Like PressurePlate itself, this isn't meant to be instantiated directly - it has no Clone() of
// its own. Concrete leaf types (StonePressurePlate, WoodenPressurePlate) embed it.
type SimplePressurePlate struct {
	PressurePlate

	Pressed bool
}

func (s *SimplePressurePlate) DescribeBlockOnlyState(w runtime.DataDescriber) { w.Bool(&s.Pressed) }

func (s *SimplePressurePlate) IsPressed() bool { return s.Pressed }

func (s *SimplePressurePlate) SetPressed(pressed bool) { s.Pressed = pressed }

func (s *SimplePressurePlate) hasOutputSignal() bool { return s.Pressed }

func (s *SimplePressurePlate) calculatePlateState(entities []Entity) (Behavior, *bool) {
	newPressed := len(entities) > 0
	if newPressed == s.Pressed {
		return s.self, nil
	}
	clone := s.self.Clone()
	clone.(interface{ SetPressed(bool) }).SetPressed(newPressed)
	return clone, &newPressed
}
