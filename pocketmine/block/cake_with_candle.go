package block

import (
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/entity/effect"
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

// GetCandle is a port of CakeWithCandle::getCandle (CakeWithDyedCandle overrides it).
func (c *CakeWithCandle) GetCandle() Behavior { return VanillaBlock("candle") }

// candle is $this->getCandle() through self, so CakeWithDyedCandle's override is used.
func (c *CakeWithCandle) candle() Behavior {
	return c.self.(interface{ GetCandle() Behavior }).GetCandle()
}

// GetResidue is a port of CakeWithCandle::getResidue.
func (c *CakeWithCandle) GetResidue() Behavior {
	cake := VanillaCake().(*Cake)
	cake.SetBites(1)
	return cake
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

// GetDropsForCompatibleTool is a port of CakeWithCandle::getDropsForCompatibleTool.
func (c *CakeWithCandle) GetDropsForCompatibleTool(item Item) []Item {
	if it := asItemOrNil(c.candle()); it != nil {
		return []Item{it}
	}
	return nil
}

// GetPickedItem is a port of CakeWithCandle::getPickedItem.
func (c *CakeWithCandle) GetPickedItem(addUserData bool) Item { return asItemOrNil(VanillaCake()) }

// OnConsume is a port of CakeWithCandle::onConsume.
func (c *CakeWithCandle) OnConsume(consumer effect.Living) {
	c.BaseCake.OnConsume(consumer)
	if world, err := c.position.GetWorld(); err == nil {
		dropItem(world, c.position.Add(0.5, 0.5, 0.5), asItemOrNil(c.candle()))
	}
}
