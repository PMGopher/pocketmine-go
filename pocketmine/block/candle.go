package block

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

const (
	candleMinCount = 1
	candleMaxCount = 4
)

// Candle is a port of pocketmine\block\Candle.
type Candle struct {
	Transparent
	CandleComponent

	Count int
}

func NewCandle(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Candle {
	c := &Candle{Transparent: Transparent{NewBlock(idInfo, name, typeInfo)}, Count: candleMinCount}
	c.Init(c)
	return c
}

func (c *Candle) Clone() Behavior {
	cl := *c
	cl.rebind(&cl)
	return &cl
}

func (c *Candle) DescribeBlockOnlyState(w runtime.DataDescriber) {
	w.Bool(&c.Lit)
	count := c.Count
	w.BoundedIntAuto(candleMinCount, candleMaxCount, &count)
	c.Count = count
}

func (c *Candle) GetCount() int { return c.Count }

func (c *Candle) SetCount(count int) {
	if count < candleMinCount || count > candleMaxCount {
		panic("Count must be in range 1 ... 4")
	}
	c.Count = count
}

func (c *Candle) GetLightLevel() int { return c.GetBaseLightLevel() * c.Count }

func (c *Candle) RecalculateCollisionBoxes() []math.AxisAlignedBB {
	var box math.AxisAlignedBB
	switch c.Count {
	case 1:
		box = math.OneAABB().SquashedCopy(math.AxisX, 7.0/16).SquashedCopy(math.AxisZ, 7.0/16)
	case 2:
		box = math.OneAABB().SquashedCopy(math.AxisX, 5.0/16).TrimmedCopy(math.North, 7.0/16).TrimmedCopy(math.South, 6.0/16)
	case 3:
		box = math.OneAABB().TrimmedCopy(math.West, 5.0/16).TrimmedCopy(math.East, 6.0/16).TrimmedCopy(math.North, 6.0/16).TrimmedCopy(math.South, 5.0/16)
	case 4:
		box = math.OneAABB().SquashedCopy(math.AxisX, 5.0/16).TrimmedCopy(math.North, 5.0/16).TrimmedCopy(math.South, 6.0/16)
	default:
		panic("Unreachable")
	}
	return []math.AxisAlignedBB{box.TrimmedCopy(math.Up, 10.0/16)}
}

func (c *Candle) GetSupportType(facing math.Facing) blockutils.SupportType {
	return blockutils.SupportTypeNone
}

func (c *Candle) getCandleIfCompatibleType(blk Behavior) *Candle {
	if candle, ok := blk.(*Candle); ok && candle.HasSameTypeId(c.self) {
		return candle
	}
	return nil
}

func (c *Candle) CanBePlacedAt(blockReplace Behavior, clickVector math.Vector3, face math.Facing, isClickedBlock bool) bool {
	if candle := c.getCandleIfCompatibleType(blockReplace); candle != nil {
		return candle.Count < candleMaxCount
	}
	return c.Transparent.CanBePlacedAt(blockReplace, clickVector, face, isClickedBlock)
}

func (c *Candle) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	if !blockReplace.(blockGeometry).GetAdjacentSupportType(math.Down).HasCenterSupport() {
		return false
	}
	if existing := c.getCandleIfCompatibleType(blockReplace); existing != nil {
		if existing.Count >= candleMaxCount {
			return false
		}
		c.Count = existing.Count + 1
		c.Lit = existing.Lit
	}
	return c.Block.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
}

func (c *Candle) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	return c.OnInteractCandle(c.self, c.position, item)
}

func (c *Candle) OnProjectileHit(projectile Projectile, hitResult math.RayTraceResult) {
	c.OnProjectileHitCandle(c.self, c.position, projectile)
}

// GetDropsForCompatibleTool should return [c.AsItem().SetCount(c.Count)] - needs real Item
// construction from the unported item package (see Block.GetDropsForCompatibleTool's doc
// comment), so this returns nil for now.
func (c *Candle) GetDropsForCompatibleTool(item Item) []Item { return nil }
