package world

import (
	"fmt"
	stdmath "math"
	"math/rand"
	"pocketmine-go/pocketmine/event"
	blockevent "pocketmine-go/pocketmine/event/block"

	"pocketmine-go/pocketmine/block"
	blockutils "pocketmine-go/pocketmine/block/utils"
	entityevent "pocketmine-go/pocketmine/event/entity"
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/world/particle"
	"pocketmine-go/pocketmine/world/sound"
	"pocketmine-go/pocketmine/world/utils"
)

// explosionRays mirrors Explosion::$rays (private int $rays = 16).
const explosionRays = 16

// Explosion is a port of pocketmine\world\Explosion.
//
// PHP's `Entity|Block|null $what` is split into two nilable fields - What (a Block source) and
// WhatEntity (an Entity source, e.g. PrimedTNT); at most one is set.
type Explosion struct {
	world      *World
	Source     math.Vector3
	Radius     float64
	What       block.Behavior
	WhatEntity Entity
	FireChance float64
	StepLen    float64

	subChunkExplorer *utils.SubChunkExplorer

	AffectedBlocks map[[3]int]block.Behavior
	affectedOrder  [][3]int
	fireIgnitions  map[[3]int]bool

	// Yield is set by ExplodeB - the drop chance (0-100) each affected block's contents roll
	// against (EntityExplodeEvent::getYield() for entity-sourced explosions).
	Yield float64
}

// NewExplosion is a port of Explosion::__construct.
func NewExplosion(source block.Position, radius float64, what block.Behavior, fireChance float64) (*Explosion, error) {
	if !source.IsValid() {
		return nil, fmt.Errorf("world: explosion source position does not have a valid world")
	}
	sourceWorld, err := source.GetWorld()
	if err != nil {
		return nil, err
	}
	w, ok := sourceWorld.(*World)
	if !ok {
		return nil, fmt.Errorf("world: explosion source's world is not a *world.World")
	}
	if stdmath.IsNaN(fireChance) || stdmath.IsInf(fireChance, 0) {
		return nil, fmt.Errorf("world: fireChance must not be NaN or infinite")
	}
	if fireChance < 0.0 || fireChance > 1.0 {
		return nil, fmt.Errorf("world: fire chance must be a number between 0 and 1, got %v", fireChance)
	}
	if radius <= 0 {
		return nil, fmt.Errorf("world: explosion radius must be greater than 0, got %v", radius)
	}

	return &Explosion{
		world:            w,
		Source:           source.AsVector3(),
		Radius:           radius,
		What:             what,
		FireChance:       fireChance,
		StepLen:          0.3,
		subChunkExplorer: utils.NewSubChunkExplorer(w),
		AffectedBlocks:   map[[3]int]block.Behavior{},
		fireIgnitions:    map[[3]int]bool{},
	}, nil
}

// NewEntityExplosion is Explosion::__construct with an Entity as $what.
func NewEntityExplosion(source block.Position, radius float64, what Entity, fireChance float64) (*Explosion, error) {
	e, err := NewExplosion(source, radius, nil, fireChance)
	if err != nil {
		return nil, err
	}
	e.WhatEntity = what
	return e, nil
}

// ExplodeA is a port of Explosion::explodeA: calculates which blocks will be destroyed by this
// explosion (ExplodeB does nothing if this hasn't been called first).
func (e *Explosion) ExplodeA() bool {
	if e.Radius < 0.1 {
		return false
	}

	mRays := explosionRays - 1
	incendiary := e.FireChance > 0

	for i := 0; i < explosionRays; i++ {
		for j := 0; j < explosionRays; j++ {
			for k := 0; k < explosionRays; k++ {
				if i != 0 && i != mRays && j != 0 && j != mRays && k != 0 && k != mRays {
					// Only the ray shell (the outer surface of the i/j/k cube) is a real ray -
					// everything else is interior and never fired, matching the PHP original.
					continue
				}

				shiftX, shiftY, shiftZ := float64(i)/float64(mRays)*2-1, float64(j)/float64(mRays)*2-1, float64(k)/float64(mRays)*2-1
				length := stdmath.Sqrt(shiftX*shiftX + shiftY*shiftY + shiftZ*shiftZ)
				shiftX, shiftY, shiftZ = (shiftX/length)*e.StepLen, (shiftY/length)*e.StepLen, (shiftZ/length)*e.StepLen

				pointerX, pointerY, pointerZ := e.Source.X, e.Source.Y, e.Source.Z

				for blastForce := e.Radius * (float64(700+rand.Intn(601)) / 1000); blastForce > 0; blastForce -= e.StepLen * 0.75 {
					x, y, z := int(pointerX), int(pointerY), int(pointerZ)
					vBlockX, vBlockY, vBlockZ := x, y, z
					if pointerX < float64(x) {
						vBlockX = x - 1
					}
					if pointerY < float64(y) {
						vBlockY = y - 1
					}
					if pointerZ < float64(z) {
						vBlockZ = z - 1
					}

					pointerX += shiftX
					pointerY += shiftY
					pointerZ += shiftZ

					if e.subChunkExplorer.MoveTo(vBlockX, vBlockY, vBlockZ) == utils.StatusInvalid {
						continue
					}
					subChunk := e.subChunkExplorer.CurrentSubChunk
					if subChunk == nil {
						continue
					}

					state := subChunk.GetBlockStateID(vBlockX&0xf, vBlockY&0xf, vBlockZ&0xf)

					blastResistance, ok := e.world.blastResistance[state]
					if !ok {
						blastResistance = 0
					}
					if blastResistance < 0 {
						continue
					}

					blastForce -= (blastResistance/5 + 0.3) * e.StepLen
					if blastForce <= 0 {
						continue
					}

					key := [3]int{vBlockX, vBlockY, vBlockZ}
					if _, already := e.AffectedBlocks[key]; already {
						continue
					}

					blk := e.world.GetBlockAt(vBlockX, vBlockY, vBlockZ)
					for _, affected := range blk.GetAffectedBlocks() {
						pos := affected.GetPosition()
						affectedKey := [3]int{pos.FloorX(), pos.FloorY(), pos.FloorZ()}
						if _, exists := e.AffectedBlocks[affectedKey]; !exists {
							e.affectedOrder = append(e.affectedOrder, affectedKey)
						}
						e.AffectedBlocks[affectedKey] = affected

						if incendiary && rand.Float64() <= e.FireChance {
							e.fireIgnitions[affectedKey] = true
						}
					}
				}
			}
		}
	}

	return true
}

// sideAccessible is the local surface Explosion's fire-placement check needs (GetSide is promoted
// from *block.Block, not part of block.Behavior itself - same "declare the exact promoted method
// this file needs" convention as World's own positionable interface).
type sideAccessible interface {
	GetSide(side math.Facing, step int) block.Behavior
}

// interceptable is the local surface getExposure's ray-blocking check needs (CalculateIntercept is
// likewise promoted from *block.Block, not part of block.Behavior).
type interceptable interface {
	CalculateIntercept(pos1, pos2 math.Vector3) (math.RayTraceResult, bool)
}

// ExplodeB is a port of Explosion::explodeB: applies the explosion's effects on the world -
// destroying blocks (if ExplodeA found any), harming and knocking back entities, and adding a
// particle and sound. See Explosion's own doc comment for what isn't ported yet (event
// cancellation and item drops).
func (e *Explosion) ExplodeB() bool {
	sourcePos := math.NewVector3(stdmath.Floor(e.Source.X), stdmath.Floor(e.Source.Y), stdmath.Floor(e.Source.Z))
	e.Yield = stdmath.Min(100, (1/e.Radius)*100)

	if e.WhatEntity != nil || e.What != nil {
		blockList := make([]entityevent.Block, 0, len(e.affectedOrder))
		for _, key := range e.affectedOrder {
			blockList = append(blockList, e.AffectedBlocks[key])
		}
		ignitions := make([]entityevent.Block, 0, len(e.fireIgnitions))
		for _, key := range e.affectedOrder {
			if e.fireIgnitions[key] {
				ignitions = append(ignitions, e.AffectedBlocks[key])
			}
		}
		position := entityevent.Position{Vector3: e.Source, World: e.world}
		var newBlocks, newIgnitions []entityevent.Block
		if e.WhatEntity != nil {
			ev := entityevent.NewEntityExplodeEvent(e.WhatEntity, position, blockList, e.Yield, ignitions)
			ev.Call()
			if ev.IsCancelled() {
				return false
			}
			e.Yield, newBlocks, newIgnitions = ev.GetYield(), ev.GetBlockList(), ev.GetIgnitions()
		} else {
			affected := make([]blockevent.Block, len(blockList))
			for i, b := range blockList {
				affected[i] = b
			}
			fire := make([]blockevent.Block, len(ignitions))
			for i, b := range ignitions {
				fire[i] = b
			}
			ev := blockevent.NewBlockExplodeEvent(e.What, position, affected, e.Yield, fire)
			event.Call(ev)
			if ev.IsCancelled() {
				return false
			}
			e.Yield = ev.GetYield()
			for _, b := range ev.GetAffectedBlocks() {
				newBlocks = append(newBlocks, b)
			}
			for _, b := range ev.GetIgnitions() {
				newIgnitions = append(newIgnitions, b)
			}
		}
		e.AffectedBlocks = map[[3]int]block.Behavior{}
		e.affectedOrder = nil
		for _, b := range newBlocks {
			if blk, ok := b.(block.Behavior); ok {
				key := positionKey(blk)
				if _, exists := e.AffectedBlocks[key]; !exists {
					e.affectedOrder = append(e.affectedOrder, key)
				}
				e.AffectedBlocks[key] = blk
			}
		}
		e.fireIgnitions = map[[3]int]bool{}
		for _, b := range newIgnitions {
			if blk, ok := b.(block.Behavior); ok {
				e.fireIgnitions[positionKey(blk)] = true
			}
		}
	}

	explosionSize := e.Radius * 2
	minX := stdmath.Floor(e.Source.X - explosionSize - 1)
	maxX := stdmath.Ceil(e.Source.X + explosionSize + 1)
	minY := stdmath.Floor(e.Source.Y - explosionSize - 1)
	maxY := stdmath.Ceil(e.Source.Y + explosionSize + 1)
	minZ := stdmath.Floor(e.Source.Z - explosionSize - 1)
	maxZ := stdmath.Ceil(e.Source.Z + explosionSize + 1)
	explosionBB := math.AxisAlignedBB{MinX: minX, MinY: minY, MinZ: minZ, MaxX: maxX, MaxY: maxY, MaxZ: maxZ}

	for _, ent := range e.world.GetNearbyEntitiesExcept(explosionBB, e.WhatEntity) {
		entPos := ent.GetPosition()
		distance := entPos.Distance(e.Source) / explosionSize

		if distance > 1 {
			continue
		}

		motion := entPos.SubtractVector(e.Source).Normalize()
		exposure := e.getExposure(e.Source, ent)
		impact := (1 - distance) * exposure
		damage := int(((impact*impact+impact)/2)*8*explosionSize + 1)

		var source entityevent.DamageSource
		if e.WhatEntity != nil {
			source = entityevent.NewEntityDamageByEntityEvent(e.WhatEntity, ent, entityevent.CauseEntityExplosion, float64(damage), nil)
		} else if e.What != nil {
			source = entityevent.NewEntityDamageByBlockEvent(e.What, ent, entityevent.CauseBlockExplosion, float64(damage), nil)
		} else {
			source = entityevent.NewEntityDamageEvent(ent, entityevent.CauseBlockExplosion, float64(damage), nil)
		}

		ent.Attack(source)
		ent.SetMotion(ent.GetMotion().AddVector(motion.Multiply(impact)))
	}

	air := block.VanillaAir()
	fire := block.VanillaFire()

	for _, key := range e.affectedOrder {
		blk := e.AffectedBlocks[key]
		x, y, z := key[0], key[1], key[2]

		if tnt, ok := blk.(*block.TNTBlock); ok {
			tnt.Ignite(10 + rand.Intn(21))
			continue
		}

		pos := block.NewPosition(float64(x), float64(y), float64(z), e.world)

		if float64(rand.Intn(101)) < e.Yield {
			air := explosionAirItem()
			var drops []block.Item
			if blk.GetBreakInfo().IsExplosionHarvestable() {
				drops = blk.GetDropsForCompatibleTool(air)
			} else {
				drops = blk.GetDrops(air)
			}
			for _, drop := range drops {
				if it, ok := drop.(item.Item); ok {
					e.world.DropItem(pos.AsVector3().Add(0.5, 0.5, 0.5), it, nil, 10)
				}
			}
		}

		if t, ok := e.world.GetTile(pos); ok {
			t.OnBlockDestroyed()
		}

		targetBlock := air
		if e.fireIgnitions[key] {
			if sa, ok := blk.(sideAccessible); ok {
				below := sa.GetSide(math.Down, 1)
				if below.GetSupportType(math.Up) == blockutils.SupportTypeFull {
					targetBlock = fire
				}
			}
		}

		_ = e.world.SetBlock(pos, targetBlock)
	}

	e.world.AddParticle(sourcePos, particle.HugeExplodeSeedParticle{})
	e.world.AddSound(sourcePos, sound.ExplodeSound{})

	return true
}

// positionKey is the [3]int block key used by AffectedBlocks/fireIgnitions.
func positionKey(blk block.Behavior) [3]int {
	pos := blk.GetPosition()
	return [3]int{pos.FloorX(), pos.FloorY(), pos.FloorZ()}
}

// explosionAirItem is PHP's VanillaItems::AIR() as passed to getDrops: an empty (null) item.
func explosionAirItem() block.Item {
	if asItem, ok := block.VanillaAir().(interface{ AsItem() (block.Item, error) }); ok {
		if it, err := asItem.AsItem(); err == nil {
			return it
		}
	}
	return nil
}

// getExposure is a port of Explosion::getExposure: the fraction of sample points across ent's
// bounding box that have an unobstructed line of sight back to origin, used to scale explosion
// impact/damage.
func (e *Explosion) getExposure(origin math.Vector3, ent block.Entity) float64 {
	bb := ent.GetBoundingBox()

	diff := math.NewVector3(bb.GetXLength(), bb.GetYLength(), bb.GetZLength()).Multiply(2).Add(1, 1, 1)
	step := math.NewVector3(1.0/diff.X, 1.0/diff.Y, 1.0/diff.Z)

	xOffset := (1.0 - (stdmath.Floor(diff.X) / diff.X)) / 2.0
	zOffset := (1.0 - (stdmath.Floor(diff.Z) / diff.Z)) / 2.0

	checks := 0.0
	hits := 0.0

	for x := 0.0; x <= 1.0; x += step.X {
		for y := 0.0; y <= 1.0; y += step.Y {
			for z := 0.0; z <= 1.0; z += step.Z {
				point := math.NewVector3(
					lerp(x, bb.MinX, bb.MaxX)+xOffset,
					lerp(y, bb.MinY, bb.MaxY),
					lerp(z, bb.MinZ, bb.MaxZ)+zOffset,
				)

				intercepted := false
				if seq, err := math.BetweenPoints(origin, point); err == nil {
					for voxel := range seq {
						vb := e.world.GetBlockAt(voxel.FloorX(), voxel.FloorY(), voxel.FloorZ())
						if ib, ok := vb.(interceptable); ok {
							if _, hit := ib.CalculateIntercept(origin, point); hit {
								intercepted = true
								break
							}
						}
					}
				}

				if !intercepted {
					hits++
				}
				checks++
			}
		}
	}

	if checks > 0 {
		return hits / checks
	}
	return 0
}

func lerp(scale, a, b float64) float64 { return a + scale*(b-a) }
