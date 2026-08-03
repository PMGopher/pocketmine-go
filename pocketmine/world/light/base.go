package light

import (
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/world/format"
	"pocketmine-go/pocketmine/world/utils"
)

// BaseLightFilter is a port of LightUpdate::BASE_LIGHT_FILTER - the default light reduction for a
// block state with no entry in LightFilters (opaque-by-default: most blocks aren't registered
// with a specific filter value, and light shouldn't pass through them for free).
const BaseLightFilter = 1

// Base is a port of the shared state and logic in the abstract pocketmine\world\light\LightUpdate
// class. SkyLightUpdate/BlockLightUpdate each embed one and supply LightUpdate's two "abstract
// method" hooks (getCurrentLightArray, and - only for SkyLightUpdate - an override of
// getEffectiveLight) as plain function fields at construction time, matching this port's
// established self-dispatch pattern for one-off polymorphism (e.g. block.Block's Init/rebind)
// since Go has no virtual-method override through struct embedding.
type Base struct {
	SubChunkExplorer *utils.SubChunkExplorer
	// LightFilters maps an internal block state ID to how much it reduces light passing through
	// it (0-15) - built by World from every registered block's real GetLightFilter(), with
	// BaseLightFilter as the fallback for anything not registered.
	LightFilters map[int32]int

	// getCurrentLightArray is LightUpdate::getCurrentLightArray, abstract in PHP.
	getCurrentLightArray func() *format.LightArray

	// effectiveLightOverride is set by SkyLightUpdate to override getEffectiveLight (its "above
	// the world = always 15" special case). Returning ok=false means "no override, use the base
	// behaviour" - matching SkyLightUpdate::getEffectiveLight's `return parent::...` fallthrough.
	effectiveLightOverride func(x, y, z int) (level int, ok bool)

	updateNodes map[pos]int
}

func newBase(explorer *utils.SubChunkExplorer, lightFilters map[int32]int) *Base {
	return &Base{SubChunkExplorer: explorer, LightFilters: lightFilters, updateNodes: map[pos]int{}}
}

func (b *Base) lightFilterFor(stateID int32) int {
	if f, ok := b.LightFilters[stateID]; ok {
		return f
	}
	return BaseLightFilter
}

// getEffectiveLight is a port of LightUpdate::getEffectiveLight.
func (b *Base) getEffectiveLight(x, y, z int) int {
	if b.effectiveLightOverride != nil {
		if level, ok := b.effectiveLightOverride(x, y, z); ok {
			return level
		}
	}
	if b.SubChunkExplorer.MoveTo(x, y, z) != utils.StatusInvalid {
		la := b.getCurrentLightArray()
		return la.Get(x&format.SubChunkCoordMask, y&format.SubChunkCoordMask, z&format.SubChunkCoordMask)
	}
	return 0
}

// getHighestAdjacentLight is a port of LightUpdate::getHighestAdjacentLight.
func (b *Base) getHighestAdjacentLight(x, y, z int) int {
	adjacent := 0
	for _, side := range math.AllFacing {
		off := math.FacingOffset[side]
		if v := b.getEffectiveLight(x+off[0], y+off[1], z+off[2]); v > adjacent {
			adjacent = v
		}
		if adjacent == 15 {
			break
		}
	}
	return adjacent
}

// SetAndUpdateLight is a port of LightUpdate::setAndUpdateLight.
func (b *Base) SetAndUpdateLight(x, y, z, newLevel int) {
	b.updateNodes[pos{x, y, z}] = newLevel
}

// prepareNodes is a port of LightUpdate::prepareNodes.
func (b *Base) prepareNodes() *PropagationContext {
	ctx := NewPropagationContext()
	for p, newLevel := range b.updateNodes {
		if b.SubChunkExplorer.MoveTo(p.X, p.Y, p.Z) == utils.StatusInvalid {
			continue
		}
		la := b.getCurrentLightArray()
		lx, ly, lz := p.X&format.SubChunkCoordMask, p.Y&format.SubChunkCoordMask, p.Z&format.SubChunkCoordMask
		oldLevel := la.Get(lx, ly, lz)

		if oldLevel == newLevel {
			continue
		}
		la.Set(lx, ly, lz, newLevel)
		if oldLevel < newLevel {
			ctx.SpreadVisited[p] = spreadFromTrue
			ctx.SpreadQueue = append(ctx.SpreadQueue, p)
		} else {
			ctx.RemovalVisited[p] = true
			ctx.RemovalQueue = append(ctx.RemovalQueue, removalEntry{p, oldLevel})
		}
	}
	return ctx
}

// Execute is a port of LightUpdate::execute: drains prepareNodes' removal queue (light that
// decreased or disappeared, propagating darkness outward and re-lighting anything that turns out
// to have another source), then its spread queue (light that increased, propagating brightness
// outward) - returns the number of positions visited, purely informational (PHP callers don't
// currently use it for anything either).
func (b *Base) Execute() int {
	ctx := b.prepareNodes()
	touched := 0

	var lightArray *format.LightArray
	b.SubChunkExplorer.Invalidate()
	for len(ctx.RemovalQueue) > 0 {
		touched++
		node := ctx.RemovalQueue[0]
		ctx.RemovalQueue = ctx.RemovalQueue[1:]

		for _, side := range math.AllFacing {
			off := math.FacingOffset[side]
			cx, cy, cz := node.X+off[0], node.Y+off[1], node.Z+off[2]

			status := b.SubChunkExplorer.MoveTo(cx, cy, cz)
			if status == utils.StatusInvalid {
				continue
			}
			if status == utils.StatusMoved {
				lightArray = b.getCurrentLightArray()
			}
			b.computeRemoveLight(cx, cy, cz, node.OldLevel, ctx, lightArray)
		}
	}

	var subChunk *format.SubChunk
	b.SubChunkExplorer.Invalidate()
	for len(ctx.SpreadQueue) > 0 {
		touched++
		p := ctx.SpreadQueue[0]
		ctx.SpreadQueue = ctx.SpreadQueue[1:]

		from := ctx.SpreadVisited[p]
		delete(ctx.SpreadVisited, p)

		status := b.SubChunkExplorer.MoveTo(p.X, p.Y, p.Z)
		if status == utils.StatusInvalid {
			continue
		}
		if status == utils.StatusMoved {
			subChunk = b.SubChunkExplorer.CurrentSubChunk
			lightArray = b.getCurrentLightArray()
		}

		newAdjacentLight := lightArray.Get(p.X&format.SubChunkCoordMask, p.Y&format.SubChunkCoordMask, p.Z&format.SubChunkCoordMask)
		if newAdjacentLight <= 0 {
			continue
		}

		for _, side := range math.AllFacing {
			if from == int(side) {
				// Don't check the side this node received its initial light from.
				continue
			}
			off := math.FacingOffset[side]
			cx, cy, cz := p.X+off[0], p.Y+off[1], p.Z+off[2]

			status := b.SubChunkExplorer.MoveTo(cx, cy, cz)
			if status == utils.StatusInvalid {
				continue
			}
			if status == utils.StatusMoved {
				subChunk = b.SubChunkExplorer.CurrentSubChunk
				lightArray = b.getCurrentLightArray()
			}
			b.computeSpreadLight(cx, cy, cz, newAdjacentLight, ctx, lightArray, subChunk, side)
		}
	}

	return touched
}

// computeRemoveLight is a port of LightUpdate::computeRemoveLight.
func (b *Base) computeRemoveLight(x, y, z, oldAdjacentLevel int, ctx *PropagationContext, lightArray *format.LightArray) {
	lx, ly, lz := x&format.SubChunkCoordMask, y&format.SubChunkCoordMask, z&format.SubChunkCoordMask
	current := lightArray.Get(lx, ly, lz)
	p := pos{x, y, z}

	if current != 0 && current < oldAdjacentLevel {
		lightArray.Set(lx, ly, lz, 0)
		if !ctx.RemovalVisited[p] {
			ctx.RemovalVisited[p] = true
			if current > 1 {
				ctx.RemovalQueue = append(ctx.RemovalQueue, removalEntry{p, current})
			}
		}
	} else if current >= oldAdjacentLevel {
		if _, visited := ctx.SpreadVisited[p]; !visited {
			ctx.SpreadVisited[p] = spreadFromTrue
			ctx.SpreadQueue = append(ctx.SpreadQueue, p)
		}
	}
}

// computeSpreadLight is a port of LightUpdate::computeSpreadLight.
func (b *Base) computeSpreadLight(x, y, z, newAdjacentLevel int, ctx *PropagationContext, lightArray *format.LightArray, subChunk *format.SubChunk, side math.Facing) {
	lx, ly, lz := x&format.SubChunkCoordMask, y&format.SubChunkCoordMask, z&format.SubChunkCoordMask
	current := lightArray.Get(lx, ly, lz)
	potentialLight := newAdjacentLevel - b.lightFilterFor(subChunk.GetBlockStateID(lx, ly, lz))

	if current < potentialLight {
		lightArray.Set(lx, ly, lz, potentialLight)
		p := pos{x, y, z}
		if _, visited := ctx.SpreadVisited[p]; !visited && potentialLight > 1 {
			// Track where this node was lit from, to avoid checking the source again when
			// propagating from here.
			ctx.SpreadVisited[p] = int(math.Opposite(side))
			ctx.SpreadQueue = append(ctx.SpreadQueue, p)
		}
	}
}
