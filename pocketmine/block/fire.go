package block

import (
	"pocketmine-go/pocketmine/event"
	blockevent "pocketmine-go/pocketmine/event/block"

	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

const FireMaxAge = 15

// Fire is a port of pocketmine\block\Fire.
type Fire struct {
	BaseFire
	AgeComponent
}

func NewFire(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Fire {
	f := &Fire{
		BaseFire:     BaseFire{Flowable{Transparent{NewBlock(idInfo, name, typeInfo)}}},
		AgeComponent: NewAgeComponent(FireMaxAge),
	}
	f.Init(f)
	return f
}

func (f *Fire) Clone() Behavior {
	c := *f
	c.rebind(&c)
	return &c
}

func (f *Fire) DescribeBlockOnlyState(w runtime.DataDescriber) { f.DescribeAge(w) }

func (f *Fire) GetFireDamage() int { return 1 }

func (f *Fire) canBeSupportedBy(blk Behavior) bool {
	return blk.GetSupportType(math.Up) == blockutils.SupportTypeFull
}

func (f *Fire) hasAdjacentFlammableBlocks() bool {
	geo := f.self.(blockGeometry)
	for _, face := range math.AllFacing {
		if geo.GetSide(face, 1).IsFlammable() {
			return true
		}
	}
	return false
}

// OnNearbyBlockChange is a port of Fire::onNearbyBlockChange.
func (f *Fire) OnNearbyBlockChange() {
	world, err := f.position.GetWorld()
	if err != nil {
		return
	}
	down := f.self.(blockGeometry).GetSide(math.Down, 1)
	if soulFireCanBeSupportedBy(down) {
		_ = world.SetBlock(f.position, VanillaBlock("soul_fire"))
	} else if !f.canBeSupportedBy(down) && !f.hasAdjacentFlammableBlocks() {
		_ = world.SetBlock(f.position, VanillaAir())
	} else {
		world.ScheduleDelayedBlockUpdate(f.position.AsVector3(), mtRand(30, 40))
	}
}

func (f *Fire) TicksRandomly() bool { return true }

// OnRandomTick is a port of Fire::onRandomTick.
func (f *Fire) OnRandomTick() {
	down := f.self.(blockGeometry).GetSide(math.Down, 1)

	var result Behavior
	if f.Age < FireMaxAge && mtRand(0, 2) == 0 {
		f.Age++
		result = f.self
	}
	canSpread := true

	if !down.BurnsForever() {
		// TODO: check rain (as in PHP)
		if f.Age == FireMaxAge {
			if !down.IsFlammable() && mtRand(0, 3) == 3 { // 1/4 chance to extinguish
				canSpread = false
				result = VanillaAir()
			}
		} else if !f.hasAdjacentFlammableBlocks() {
			canSpread = false
			if down.IsTransparent() || f.Age > 3 {
				result = VanillaAir()
			}
		}
	}

	world, err := f.position.GetWorld()
	if err != nil {
		return
	}
	if result != nil {
		_ = world.SetBlock(f.position, result)
	}

	world.ScheduleDelayedBlockUpdate(f.position.AsVector3(), mtRand(30, 40))

	if canSpread {
		f.burnBlocksAround()
		f.spreadFire()
	}
}

func (f *Fire) OnScheduledUpdate() { f.self.OnRandomTick() }

// burnBlocksAround is a port of Fire::burnBlocksAround.
func (f *Fire) burnBlocksAround() {
	geo := f.self.(blockGeometry)
	// TODO: raise upper bound for chance in humid biomes (as in PHP)
	for _, face := range horizontalFacings {
		f.burnBlock(geo.GetSide(face, 1), 300)
	}
	// vanilla uses a 250 upper bound here, but PHP doesn't think they intended to increase the
	// chance of incineration
	f.burnBlock(geo.GetSide(math.Up, 1), 350)
	f.burnBlock(geo.GetSide(math.Down, 1), 350)
}

// burnBlock is a port of Fire::burnBlock.
func (f *Fire) burnBlock(blk Behavior, chanceBound int) {
	if mtRand(0, chanceBound) >= blk.GetFlammability() {
		return
	}
	ev := blockevent.NewBlockBurnEvent(blk, f.self)
	event.Call(ev)
	if ev.IsCancelled() {
		return
	}
	blk.OnIncinerate()

	world, err := f.position.GetWorld()
	if err != nil {
		return
	}
	pos := blk.GetPosition()
	if world.GetBlockAt(pos.FloorX(), pos.FloorY(), pos.FloorZ()).GetStateId() != blk.GetStateId() {
		return
	}
	spreadedFire := false
	if mtRand(0, f.Age+9) < 5 { // TODO: check rain (as in PHP)
		fire := f.self.Clone().(*Fire)
		fire.Age = min(FireMaxAge, fire.Age+(mtRand(0, 4)>>2))
		spreadedFire = Spread(blk, fire, f.self)
	}
	if !spreadedFire {
		_ = world.SetBlock(pos, VanillaAir())
	}
}

// spreadFire is a port of Fire::spreadFire. PHP's World::Y_MIN/Y_MAX check on targetY is covered
// by the isInWorld check below (this package can't import world for the constants).
func (f *Fire) spreadFire() {
	world, err := f.position.GetWorld()
	if err != nil {
		return
	}
	difficulty := 0
	if d, ok := world.(interface{ GetDifficulty() int }); ok {
		difficulty = d.GetDifficulty()
	}
	chunkLoaded := func(chunkX, chunkZ int) bool {
		if c, ok := world.(interface{ IsChunkLoaded(chunkX, chunkZ int) bool }); ok {
			return c.IsChunkLoaded(chunkX, chunkZ)
		}
		return true
	}
	difficultyChanceIncrease := difficulty * 7
	ageDivisor := f.Age + 30
	baseX, baseY, baseZ := f.position.FloorX(), f.position.FloorY(), f.position.FloorZ()

	for y := -1; y <= 4; y++ {
		targetY := y + baseY
		// Higher blocks have a lower chance of catching fire
		randomBound := 100
		if y > 1 {
			randomBound += (y - 1) * 100
		}
		for z := -1; z <= 1; z++ {
			targetZ := z + baseZ
			for x := -1; x <= 1; x++ {
				if x == 0 && y == 0 && z == 0 {
					continue
				}
				targetX := x + baseX
				if !world.IsInWorld(targetX, targetY, targetZ) || !chunkLoaded(targetX>>4, targetZ>>4) {
					continue
				}
				blk := world.GetBlockAt(targetX, targetY, targetZ)
				if blk.GetTypeId() != AIR {
					continue
				}

				// TODO: fire can't spread if it's raining in any horizontally adjacent block, or the
				// current one (as in PHP)

				encouragement := 0
				for _, side := range math.NewVector3(float64(targetX), float64(targetY), float64(targetZ)).Sides(1) {
					sx, sy, sz := side.FloorX(), side.FloorY(), side.FloorZ()
					if world.IsInWorld(sx, sy, sz) {
						encouragement = max(encouragement, world.GetBlockAt(sx, sy, sz).GetFlameEncouragement())
					}
				}
				if encouragement <= 0 {
					continue
				}

				maxChance := (encouragement + 40 + difficultyChanceIncrease) / ageDivisor
				// TODO: max chance is lowered by half in humid biomes (as in PHP)

				if maxChance > 0 && mtRand(0, randomBound-1) <= maxChance {
					fire := f.self.Clone().(*Fire)
					fire.Age = min(FireMaxAge, f.Age+(mtRand(0, 4)>>2))
					Spread(blk, fire, f.self)
				}
			}
		}
	}
}
