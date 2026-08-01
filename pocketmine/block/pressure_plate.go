package block

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/world/sound"
)

// pressurePlateShaper is implemented by concrete/intermediate pressure plate types
// (SimplePressurePlate, WeightedPressurePlate) to plug into the shared PressurePlate logic below -
// same self-dispatch need and reasoning as RailShaper for the rail subsystem. Unexported since,
// like the PHP originals' `protected`, only code within this package needs it.
type pressurePlateShaper interface {
	Behavior
	hasOutputSignal() bool
	// calculatePlateState returns the new block state to write (or the same Behavior if
	// unchanged) and, if the pressed/active status changed, a non-nil pressedChange.
	calculatePlateState(entities []Entity) (newState Behavior, pressedChange *bool)
	filterIrrelevantEntities(entities []Entity) []Entity
}

// PressurePlate is a port of pocketmine\block\PressurePlate.
//
// Like Button/Crops, this isn't meant to be instantiated directly - it has no Clone() of its own.
type PressurePlate struct {
	Transparent

	DeactivationDelayTicks int
}

func (p *PressurePlate) IsSolid() bool { return false }

func (p *PressurePlate) RecalculateCollisionBoxes() []math.AxisAlignedBB { return nil }

func (p *PressurePlate) GetSupportType(facing math.Facing) blockutils.SupportType {
	return blockutils.SupportTypeNone
}

func (p *PressurePlate) canBeSupportedAt(blk Behavior) bool {
	return blk.(blockGeometry).GetAdjacentSupportType(math.Down) != blockutils.SupportTypeNone
}

func (p *PressurePlate) CanBePlacedAt(blockReplace Behavior, clickVector math.Vector3, face math.Facing, isClickedBlock bool) bool {
	return p.canBeSupportedAt(blockReplace) && p.Transparent.CanBePlacedAt(blockReplace, clickVector, face, isClickedBlock)
}

func (p *PressurePlate) OnNearbyBlockChange() {
	if !p.canBeSupportedAt(p.self) {
		if world, err := p.position.GetWorld(); err == nil {
			world.UseBreakOn(p.position.AsVector3())
		}
	} else {
		p.Transparent.OnNearbyBlockChange()
	}
}

func (p *PressurePlate) HasEntityCollision() bool { return true }

func (p *PressurePlate) OnEntityInside(entity Entity) bool {
	if !p.self.(pressurePlateShaper).hasOutputSignal() {
		if world, err := p.position.GetWorld(); err == nil {
			world.ScheduleDelayedBlockUpdate(p.position.AsVector3(), 0)
		}
	}
	return true
}

// getActivationBox is the AABB entities must intersect to activate the pressure plate - not the
// same as the (empty) collision box, nor the visual bounding box. It has a height of 0.25 blocks.
func (p *PressurePlate) getActivationBox() math.AxisAlignedBB {
	pos := p.position.AsVector3()
	return math.OneAABB().
		SquashedCopy(math.AxisX, 1.0/8).
		SquashedCopy(math.AxisZ, 1.0/8).
		TrimmedCopy(math.Up, 3.0/4).
		OffsetCopy(pos.X, pos.Y, pos.Z)
}

// hasOutputSignal/calculatePlateState/filterIrrelevantEntities are TODOs in the PHP original too
// ("make this abstract in PM6") - these are the defaults, overridden by SimplePressurePlate and
// WeightedPressurePlate.
func (p *PressurePlate) hasOutputSignal() bool { return false }

func (p *PressurePlate) calculatePlateState(entities []Entity) (Behavior, *bool) {
	return p.self, nil
}

func (p *PressurePlate) filterIrrelevantEntities(entities []Entity) []Entity { return entities }

func (p *PressurePlate) OnScheduledUpdate() {
	world, err := p.position.GetWorld()
	if err != nil {
		return
	}
	self := p.self.(pressurePlateShaper)

	activatingEntities := self.filterIrrelevantEntities(world.GetNearbyEntities(p.getActivationBox()))

	// If an irrelevant entity is inside the full cube space of the pressure plate but not
	// activating the plate, it will cause scheduled updates on the plate every tick. We don't
	// want to fire events in this case if the plate is already deactivated.
	if len(activatingEntities) == 0 && !self.hasOutputSignal() {
		return
	}

	newState, pressedChange := self.calculatePlateState(activatingEntities)

	// PressurePlateUpdateEvent isn't fired (deferred concrete event subclass - see the project
	// todo list), so newState/pressedChange are always taken as-is (never cancelled).

	if newState != nil {
		if err := world.SetBlock(p.position, newState); err != nil {
			panic(err)
		}
		if pressedChange != nil {
			var s sound.Sound
			if *pressedChange {
				s = sound.PressurePlateActivateSound{BlockStateID: newState.GetStateId()}
			} else {
				s = sound.PressurePlateDeactivateSound{BlockStateID: newState.GetStateId()}
			}
			world.AddSound(p.position.AsVector3(), s)
		}
	}

	shouldSchedule := self.hasOutputSignal()
	if pressedChange != nil {
		shouldSchedule = *pressedChange
	}
	if shouldSchedule {
		world.ScheduleDelayedBlockUpdate(p.position.AsVector3(), p.DeactivationDelayTicks)
	}
}
