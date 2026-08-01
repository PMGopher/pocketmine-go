package block

import (
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

const cakeMaxBites = 6

// ItemBlockLike is a forward-compatible marker for pocketmine\item\ItemBlock - same pattern as
// the Dye interface in base_sign.go.
type ItemBlockLike interface {
	GetBlock() Behavior
}

// Cake is a port of pocketmine\block\Cake.
type Cake struct {
	BaseCake

	Bites int
}

func NewCake(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Cake {
	c := &Cake{BaseCake: BaseCake{Transparent{NewBlock(idInfo, name, typeInfo)}}}
	c.Init(c)
	return c
}

func (c *Cake) Clone() Behavior {
	cl := *c
	cl.rebind(&cl)
	return &cl
}

func (c *Cake) DescribeBlockOnlyState(w runtime.DataDescriber) {
	w.BoundedIntAuto(0, cakeMaxBites, &c.Bites)
}

func (c *Cake) RecalculateCollisionBoxes() []math.AxisAlignedBB {
	return []math.AxisAlignedBB{
		math.OneAABB().
			ContractedCopy(1.0/16, 0, 1.0/16).
			TrimmedCopy(math.Up, 0.5).
			TrimmedCopy(math.West, float64(c.Bites)/8),
	}
}

func (c *Cake) GetBites() int { return c.Bites }

func (c *Cake) SetBites(bites int) {
	if bites < 0 || bites > cakeMaxBites {
		panic("Bites must be in range 0 ... 6")
	}
	c.Bites = bites
}

// OnInteract is a port of Cake::onInteract. The candle-topping branch (turning a bite-free cake
// into CakeWithCandle/CakeWithDyedCandle when clicked with a candle ItemBlock) needs the unported
// block registry (VanillaBlocks) to construct the result block, so it's skipped and this falls
// straight through to BaseCake.OnInteract, same gap as Farmland/GrassPath's VanillaBlocks swaps.
func (c *Cake) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	return c.BaseCake.OnInteract(item, face, clickVector, player, returnedItems)
}

func (c *Cake) GetDropsForCompatibleTool(item Item) []Item { return nil }

// GetResidue should return a clone with Bites incremented, or VanillaBlocks.AIR() once bites
// exceeds the max - needs the unported block registry, same gap documented on OnInteract above.
func (c *Cake) GetResidue() Behavior { return c.self }
