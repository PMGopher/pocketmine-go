package block

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/event"
	blockevent "pocketmine-go/pocketmine/event/block"
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

	flowVector *math.Vector3
}

func (l *Liquid) liquidBase() *Liquid { return l }

// IsLiquid reports whether blk is a Liquid (Water, Lava, ...) - an exported `instanceof Liquid`
// check for packages outside block that can't reach the unexported liquidBaser interface
// themselves (e.g. populator/groundcover.GroundCover's port of `$id instanceof Liquid`).
func IsLiquid(blk Behavior) bool {
	_, ok := blk.(liquidBaser)
	return ok
}

// liquidShaper lets concrete leaf types (Water, Lava) provide their own tick rate and override
// checkForHarden, reached from Liquid's own OnNearbyBlockChange/OnScheduledUpdate via self-dispatch
// - same self-dispatch shape as every other *Shaper in this port.
type liquidShaper interface {
	TickRate() int
	checkForHarden() bool
	GetFlowDecayPerBlock() int
	GetMinAdjacentSourcesToFormSource() (int, bool)
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

// getEffectiveFlowDecay is a port of Liquid::getEffectiveFlowDecay: -1 if blk isn't this liquid.
func (l *Liquid) getEffectiveFlowDecay(blk Behavior) int {
	lb, ok := blk.(liquidBaser)
	if !ok || blk.GetTypeId() != l.self.GetTypeId() {
		return -1
	}
	other := lb.liquidBase()
	if other.Falling {
		return 0
	}
	return other.Decay
}

// ReadStateFromWorld is a port of Liquid::readStateFromWorld.
func (l *Liquid) ReadStateFromWorld() Behavior {
	l.Block.ReadStateFromWorld()
	l.flowVector = nil
	return l.self
}

// OnScheduledUpdate is a port of Liquid::onScheduledUpdate: the flow/spread algorithm.
func (l *Liquid) OnScheduledUpdate() {
	shaper := l.self.(liquidShaper)
	multiplier := shaper.GetFlowDecayPerBlock()
	world, err := l.position.GetWorld()
	if err != nil {
		return
	}
	x, y, z := l.position.FloorX(), l.position.FloorY(), l.position.FloorZ()

	if !l.IsSource() {
		smallestFlowDecay := -100
		l.AdjacentSources = 0
		smallestFlowDecay = l.getSmallestFlowDecay(world.GetBlockAt(x, y, z-1), smallestFlowDecay)
		smallestFlowDecay = l.getSmallestFlowDecay(world.GetBlockAt(x, y, z+1), smallestFlowDecay)
		smallestFlowDecay = l.getSmallestFlowDecay(world.GetBlockAt(x-1, y, z), smallestFlowDecay)
		smallestFlowDecay = l.getSmallestFlowDecay(world.GetBlockAt(x+1, y, z), smallestFlowDecay)

		newDecay := smallestFlowDecay + multiplier
		falling := false

		if newDecay > LiquidMaxDecay || smallestFlowDecay < 0 {
			newDecay = -1
		}
		if l.getEffectiveFlowDecay(world.GetBlockAt(x, y+1, z)) >= 0 {
			falling = true
		}

		if minAdjacentSources, ok := shaper.GetMinAdjacentSourcesToFormSource(); ok && l.AdjacentSources >= minAdjacentSources {
			bottomBlock := world.GetBlockAt(x, y-1, z)
			bottomLiquid, isLiquid := bottomBlock.(liquidBaser)
			if bottomBlock.IsSolid() || (isLiquid && bottomBlock.GetTypeId() == l.self.GetTypeId() && bottomLiquid.liquidBase().IsSource()) {
				newDecay = 0
				falling = false
			}
		}

		if falling != l.Falling || (!falling && newDecay != l.Decay) {
			if !falling && newDecay < 0 {
				_ = world.SetBlock(l.position, VanillaAir())
				return
			}
			l.Falling = falling
			l.Decay = newDecay
			if falling {
				l.Decay = 0
			}
			_ = world.SetBlock(l.position, l.self) // local block update will cause an update to be scheduled
		}
	}

	bottomBlock := world.GetBlockAt(x, y-1, z)
	l.flowIntoBlock(bottomBlock, 0, true)

	if l.IsSource() || !bottomBlock.CanBeFlowedInto() {
		adjacentDecay := l.Decay + multiplier
		if l.Falling {
			adjacentDecay = 1 // falling liquid behaves like source block
		}
		if adjacentDecay <= LiquidMaxDecay {
			calculator := NewMinimumCostFlowCalculator(world, shaper.GetFlowDecayPerBlock(), l.canFlowInto)
			for _, facing := range calculator.GetOptimalFlowDirections(x, y, z) {
				offset := math.FacingOffset[facing]
				l.flowIntoBlock(world.GetBlockAt(x+offset[0], y+offset[1], z+offset[2]), adjacentDecay, false)
			}
		}
	}

	shaper.checkForHarden()
}

// flowIntoBlock is a port of Liquid::flowIntoBlock.
func (l *Liquid) flowIntoBlock(blk Behavior, newFlowDecay int, falling bool) {
	if _, isLiquid := blk.(liquidBaser); !l.canFlowInto(blk) || isLiquid {
		return
	}
	newState := l.self.Clone()
	nl := newState.(liquidBaser).liquidBase()
	nl.Falling = falling
	nl.Decay = newFlowDecay
	if falling {
		nl.Decay = 0
	}
	ev := blockevent.NewBlockSpreadEvent(blk, l.self, newState)
	event.Call(ev)
	if ev.IsCancelled() {
		return
	}
	world, err := l.position.GetWorld()
	if err != nil {
		return
	}
	if blk.GetTypeId() != AIR {
		world.UseBreakOn(blk.GetPosition().Vector3)
	}
	_ = world.SetBlock(blk.GetPosition(), ev.GetNewState().(Behavior))
}

// getSmallestFlowDecay is a port of Liquid::getSmallestFlowDecay.
func (l *Liquid) getSmallestFlowDecay(blk Behavior, decay int) int {
	lb, ok := blk.(liquidBaser)
	if !ok || blk.GetTypeId() != l.self.GetTypeId() {
		return decay
	}
	other := lb.liquidBase()
	blockDecay := other.Decay
	if other.IsSource() {
		l.AdjacentSources++
	} else if other.Falling {
		blockDecay = 0
	}
	if decay >= 0 && blockDecay >= decay {
		return decay
	}
	return blockDecay
}

// GetFlowVector is a port of Liquid::getFlowVector: the direction the liquid pushes entities in.
func (l *Liquid) GetFlowVector() math.Vector3 {
	if l.flowVector != nil {
		return *l.flowVector
	}
	world, err := l.position.GetWorld()
	if err != nil {
		return math.Vector3{}
	}
	vX, vY, vZ := 0, 0, 0
	x, y, z := l.position.FloorX(), l.position.FloorY(), l.position.FloorZ()
	decay := l.getEffectiveFlowDecay(l.self)

	for _, j := range horizontalFacings {
		offset := math.FacingOffset[j]
		dx, dy, dz := offset[0], offset[1], offset[2]
		sideX, sideY, sideZ := x+dx, y+dy, z+dz
		sideBlock := world.GetBlockAt(sideX, sideY, sideZ)
		blockDecay := l.getEffectiveFlowDecay(sideBlock)

		if blockDecay < 0 {
			if !sideBlock.CanBeFlowedInto() {
				continue
			}
			blockDecay = l.getEffectiveFlowDecay(world.GetBlockAt(sideX, sideY-1, sideZ))
			if blockDecay >= 0 {
				realDecay := blockDecay - (decay - 8)
				vX += dx * realDecay
				vY += dy * realDecay
				vZ += dz * realDecay
			}
			continue
		}
		realDecay := blockDecay - decay
		vX += dx * realDecay
		vY += dy * realDecay
		vZ += dz * realDecay
	}

	vector := math.NewVector3(float64(vX), float64(vY), float64(vZ))
	if l.Falling {
		for _, facing := range horizontalFacings {
			offset := math.FacingOffset[facing]
			if !l.canFlowInto(world.GetBlockAt(x+offset[0], y+offset[1], z+offset[2])) ||
				!l.canFlowInto(world.GetBlockAt(x+offset[0], y+offset[1]+1, z+offset[2])) {
				vector = vector.Normalize().Add(0, -6, 0)
				break
			}
		}
	}
	normalized := vector.Normalize()
	l.flowVector = &normalized
	return normalized
}

func (l *Liquid) AddVelocityToEntity(entity Entity) (math.Vector3, bool) {
	if entity.CanBeMovedByCurrents() {
		return l.GetFlowVector(), true
	}
	return math.Vector3{}, false
}
