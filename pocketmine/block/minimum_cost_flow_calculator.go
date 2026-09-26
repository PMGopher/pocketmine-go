package block

import "pocketmine-go/pocketmine/math"

// MinimumCostFlowCalculator status values.
const (
	flowCanFlowDown = 1
	flowCanFlow     = 0
	flowBlocked     = -1
)

// horizontalFacings is Facing::HORIZONTAL, in PHP's order.
var horizontalFacings = []math.Facing{math.North, math.South, math.West, math.East}

// MinimumCostFlowCalculator is a port of pocketmine\block\utils\MinimumCostFlowCalculator (in
// this package, like BlockEventHelper, since it works on blocks and the block World interface):
// the horizontal directions in which a liquid should flow to reach the nearest drop.
type MinimumCostFlowCalculator struct {
	world             World
	flowDecayPerBlock int
	canFlowInto       func(blk Behavior) bool
	flowCostVisited   map[[3]int]int
}

// NewMinimumCostFlowCalculator is a port of MinimumCostFlowCalculator::__construct.
func NewMinimumCostFlowCalculator(world World, flowDecayPerBlock int, canFlowInto func(blk Behavior) bool) *MinimumCostFlowCalculator {
	return &MinimumCostFlowCalculator{world: world, flowDecayPerBlock: flowDecayPerBlock, canFlowInto: canFlowInto, flowCostVisited: map[[3]int]int{}}
}

func (c *MinimumCostFlowCalculator) status(x, y, z int) int {
	if !c.world.IsInWorld(x, y, z) || !c.canFlowInto(c.world.GetBlockAt(x, y, z)) {
		return flowBlocked
	}
	if c.world.GetBlockAt(x, y-1, z).CanBeFlowedInto() {
		return flowCanFlowDown
	}
	return flowCanFlow
}

// calculateFlowCost is a port of MinimumCostFlowCalculator::calculateFlowCost.
func (c *MinimumCostFlowCalculator) calculateFlowCost(blockX, blockY, blockZ, accumulatedCost, maxCost int, originOpposite, lastOpposite math.Facing) int {
	cost := 1000
	for _, j := range horizontalFacings {
		if j == originOpposite || j == lastOpposite {
			continue
		}
		offset := math.FacingOffset[j]
		x, y, z := blockX+offset[0], blockY+offset[1], blockZ+offset[2]
		key := [3]int{x, y, z}
		status, visited := c.flowCostVisited[key]
		if !visited {
			status = c.status(x, y, z)
			c.flowCostVisited[key] = status
		}
		if status == flowBlocked {
			continue
		} else if status == flowCanFlowDown {
			return accumulatedCost
		}
		if accumulatedCost >= maxCost {
			continue
		}
		if realCost := c.calculateFlowCost(x, y, z, accumulatedCost+1, maxCost, originOpposite, math.Opposite(j)); realCost < cost {
			cost = realCost
		}
	}
	return cost
}

// GetOptimalFlowDirections is a port of MinimumCostFlowCalculator::getOptimalFlowDirections.
func (c *MinimumCostFlowCalculator) GetOptimalFlowDirections(originX, originY, originZ int) []math.Facing {
	flowCost := map[math.Facing]int{}
	for _, j := range horizontalFacings {
		flowCost[j] = 1000
	}
	maxCost := 4 / c.flowDecayPerBlock
	for _, j := range horizontalFacings {
		offset := math.FacingOffset[j]
		x, y, z := originX+offset[0], originY+offset[1], originZ+offset[2]
		key := [3]int{x, y, z}
		if !c.world.IsInWorld(x, y, z) || !c.canFlowInto(c.world.GetBlockAt(x, y, z)) {
			c.flowCostVisited[key] = flowBlocked
		} else if c.world.GetBlockAt(x, y-1, z).CanBeFlowedInto() {
			c.flowCostVisited[key] = flowCanFlowDown
			flowCost[j] = 0
			maxCost = 0
		} else if maxCost > 0 {
			c.flowCostVisited[key] = flowCanFlow
			opposite := math.Opposite(j)
			flowCost[j] = c.calculateFlowCost(x, y, z, 1, maxCost, opposite, opposite)
			maxCost = min(maxCost, flowCost[j])
		}
	}
	c.flowCostVisited = map[[3]int]int{}

	minCost := 1 << 30
	for _, j := range horizontalFacings {
		minCost = min(minCost, flowCost[j])
	}
	var optimal []math.Facing
	for _, j := range horizontalFacings {
		if flowCost[j] == minCost {
			optimal = append(optimal, j)
		}
	}
	return optimal
}
