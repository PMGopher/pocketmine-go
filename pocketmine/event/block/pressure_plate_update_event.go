package block

// PressurePlateUpdateEvent is a port of pocketmine\event\block\PressurePlateUpdateEvent: called
// whenever the list of entities on a pressure plate changes. Depending on the type of pressure
// plate, this might turn on/off its signal, or change the signal strength.
type PressurePlateUpdateEvent struct {
	BaseBlockChangeEvent

	activatingEntities []Entity
}

func NewPressurePlateUpdateEvent(block, newState Block, activatingEntities []Entity) *PressurePlateUpdateEvent {
	return &PressurePlateUpdateEvent{BaseBlockChangeEvent: NewBaseBlockChangeEventBase(block, newState), activatingEntities: activatingEntities}
}

// GetActivatingEntities returns a list of entities intersecting the pressure plate's activation
// box. If the pressure plate is about to deactivate, this list will be empty.
func (e *PressurePlateUpdateEvent) GetActivatingEntities() []Entity { return e.activatingEntities }
