package block

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/utils"
	"pocketmine-go/pocketmine/world/sound"
)

const LiquidMaxDecay = 7

// liquidBaser is satisfied by Liquid and (via embedding) anything that embeds it, such as Water
// or Lava - used so a neighbouring *Water/*Lava is recognised as "a Liquid" the way PHP's
// `instanceof Liquid` does for a subclass, instead of only matching an exact concrete type. Same
// pattern as candleBaser/fenceBaser/thinBaser. Replaces the temporary Liquid marker interface
// that used to live in flowable.go, now that the real type exists (same cleanup already done once
// for FenceGate/Thin in wall.go).
type liquidBaser interface {
	liquidBase() *Liquid
}

// Liquid is a port of pocketmine\block\Liquid. Like Crops/Stem, this isn't meant to be
// instantiated directly - a concrete leaf type (Water, Lava) must embed it, implement Clone, and
// provide TickRate/GetBucketFillSound/GetBucketEmptySound.
type Liquid struct {
	Transparent

	AdjacentSources int
	Falling         bool
	Decay           int
	Still           bool
}

func (l *Liquid) liquidBase() *Liquid { return l }

// liquidShaper lets concrete leaf types (Water, Lava) provide their own tick rate and override
// checkForHarden, reached from Liquid's own OnNearbyBlockChange/OnScheduledUpdate via self-dispatch
// - same self-dispatch shape as every other *Shaper in this port.
type liquidShaper interface {
	TickRate() int
	checkForHarden() bool
}

func (l *Liquid) DescribeBlockOnlyState(w runtime.DataDescriber) {
	decay := l.Decay
	w.BoundedIntAuto(0, LiquidMaxDecay, &decay)
	l.Decay = decay
	w.Bool(&l.Falling)
	w.Bool(&l.Still)
}

func (l *Liquid) IsFalling() bool { return l.Falling }

func (l *Liquid) SetFalling(falling bool) { l.Falling = falling }

func (l *Liquid) GetDecay() int { return l.Decay }

// SetDecay panics if out of range, mirroring the PHP original's InvalidArgumentException (a
// programmer error at the call site).
func (l *Liquid) SetDecay(decay int) {
	if decay < 0 || decay > LiquidMaxDecay {
		panic("Decay must be in range 0 ... 7")
	}
	l.Decay = decay
}

func (l *Liquid) HasEntityCollision() bool { return true }

func (l *Liquid) CanBeReplaced() bool { return true }

func (l *Liquid) CanBeFlowedInto() bool { return true }

func (l *Liquid) IsSolid() bool { return false }

func (l *Liquid) RecalculateCollisionBoxes() []math.AxisAlignedBB { return nil }

func (l *Liquid) GetSupportType(facing math.Facing) blockutils.SupportType {
	return blockutils.SupportTypeNone
}

// GetDropsForCompatibleTool deliberately returns nothing, matching the PHP original's
// `return [];` (this isn't a not-yet-ported gap).
func (l *Liquid) GetDropsForCompatibleTool(item Item) []Item { return nil }

// GetStillForm/GetFlowingForm are ports of Liquid::getStillForm/getFlowingForm.
func (l *Liquid) GetStillForm() Behavior {
	clone := l.self.Clone()
	clone.(liquidBaser).liquidBase().Still = true
	return clone
}

func (l *Liquid) GetFlowingForm() Behavior {
	clone := l.self.Clone()
	clone.(liquidBaser).liquidBase().Still = false
	return clone
}

func (l *Liquid) IsSource() bool { return !l.Falling && l.Decay == 0 }

func (l *Liquid) GetFluidHeightPercent() float64 {
	d := l.Decay
	if l.Falling {
		d = 0
	}
	return float64(d+1) / 9
}

func (l *Liquid) IsStill() bool { return l.Still }

func (l *Liquid) SetStill(still bool) { l.Still = still }

func (l *Liquid) GetFlowDecayPerBlock() int { return 1 }

// GetMinAdjacentSourcesToFormSource returns the number of horizontally-adjacent source blocks
// needed for this block to become a source itself, and false if this liquid doesn't exhibit
// source-forming behaviour (matching the PHP original's nullable int return).
func (l *Liquid) GetMinAdjacentSourcesToFormSource() (int, bool) { return 0, false }

// checkForHarden is a port of Liquid::checkForHarden's default (false) - Lava overrides this (see
// lava.go).
func (l *Liquid) checkForHarden() bool { return false }

// liquidCollide is a port of Liquid::liquidCollide.
func (l *Liquid) liquidCollide(cause Behavior, result Behavior) bool {
	if Form(l.self, result, cause) {
		if world, err := l.position.GetWorld(); err == nil {
			pitch := 2.6 + (utils.GetRandomFloat()-utils.GetRandomFloat())*0.8
			world.AddSound(l.position.AsVector3().Add(0.5, 0.5, 0.5), sound.FizzSound{Pitch: pitch})
		}
	}
	return true
}

// canFlowInto is a port of Liquid::canFlowInto.
func (l *Liquid) canFlowInto(blk Behavior) bool {
	world, err := l.position.GetWorld()
	if err != nil {
		return false
	}
	pos := blk.GetPosition()
	if !world.IsInWorld(pos.FloorX(), pos.FloorY(), pos.FloorZ()) {
		return false
	}
	if !blk.CanBeFlowedInto() {
		return false
	}
	if lb, ok := blk.(liquidBaser); ok && lb.liquidBase().IsSource() {
		return false
	}
	return true
}

// OnNearbyBlockChange is a port of Liquid::onNearbyBlockChange.
func (l *Liquid) OnNearbyBlockChange() {
	shaper := l.self.(liquidShaper)
	if !shaper.checkForHarden() {
		if world, err := l.position.GetWorld(); err == nil {
			world.ScheduleDelayedBlockUpdate(l.position.AsVector3(), shaper.TickRate())
		}
	}
}

// OnScheduledUpdate is a port of Liquid::onScheduledUpdate - the actual flow/spread algorithm.
// It needs MinimumCostFlowCalculator (a whole pathfinding helper class), BlockSpreadEvent, and the
// block registry (VanillaBlocks.AIR() as a replacement state), none ported yet, so this is a
// documented no-op for now.
func (l *Liquid) OnScheduledUpdate() {}

// GetFlowVector is a port of Liquid::getFlowVector. The real algorithm inspects neighbouring
// liquid decay levels via World.GetBlockAt in all four horizontal directions (already reachable),
// but its result only ever feeds AddVelocityToEntity, and there's no entity-physics subsystem yet
// to consume that vector, so this is a documented no-op returning a zero vector for now.
func (l *Liquid) GetFlowVector() math.Vector3 { return math.Vector3{} }

func (l *Liquid) AddVelocityToEntity(entity Entity) (math.Vector3, bool) {
	if entity.CanBeMovedByCurrents() {
		return l.GetFlowVector(), true
	}
	return math.Vector3{}, false
}
