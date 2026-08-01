package block

import (
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// CakeWithCandle is a port of pocketmine\block\CakeWithCandle.
type CakeWithCandle struct {
	BaseCake
	CandleComponent
}

func NewCakeWithCandle(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *CakeWithCandle {
	c := &CakeWithCandle{BaseCake: BaseCake{Transparent{NewBlock(idInfo, name, typeInfo)}}}
	c.Init(c)
	return c
}

func (c *CakeWithCandle) Clone() Behavior {
	cl := *c
	cl.rebind(&cl)
	return &cl
}

func (c *CakeWithCandle) DescribeBlockOnlyState(w runtime.DataDescriber) { c.DescribeLit(w) }

// GetLightLevel is a port of CandleTrait::getLightLevel — unlike Candle, CakeWithCandle has no
// stack Count to multiply by.
func (c *CakeWithCandle) GetLightLevel() int { return c.GetBaseLightLevel() }

func (c *CakeWithCandle) OnProjectileHit(projectile Projectile, hitResult math.RayTraceResult) {
	c.OnProjectileHitCandle(c.self, c.position, projectile)
}

func (c *CakeWithCandle) RecalculateCollisionBoxes() []math.AxisAlignedBB {
	return []math.AxisAlignedBB{
		math.OneAABB().
			ContractedCopy(1.0/16, 0, 1.0/16).
			TrimmedCopy(math.Up, 0.5), // TODO: not sure if the candle affects height
	}
}

// GetCandle should return VanillaBlocks.CANDLE() - needs the unported block registry, so this
// returns a bare, unpositioned Candle instead (enough for GetDropsForCompatibleTool below, which
// only reads its type through AsItem's still-unported machinery anyway).
func (c *CakeWithCandle) GetCandle() *Candle {
	return &Candle{Count: candleMinCount}
}

// OnInteract is a port of CakeWithCandle::onInteract.
func (c *CakeWithCandle) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	if c.Lit && face != math.Up {
		return true
	}
	if c.OnInteractCandle(c.self, c.position, item) {
		return true
	}
	return c.BaseCake.OnInteract(item, face, clickVector, player, returnedItems)
}

func (c *CakeWithCandle) GetDropsForCompatibleTool(item Item) []Item { return nil }

// OnConsume is a port of CakeWithCandle::onConsume, minus the residue swap (see BaseCake.OnConsume's
// doc comment) and the world.dropItem(candle) call, which needs real Item construction from the
// unported item package.
func (c *CakeWithCandle) OnConsume(consumer Living) { c.BaseCake.OnConsume(consumer) }
