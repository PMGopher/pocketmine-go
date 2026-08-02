package block

// Cobweb is a port of pocketmine\block\Cobweb.
type Cobweb struct {
	Flowable
}

func NewCobweb(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Cobweb {
	c := &Cobweb{Flowable: Flowable{Transparent{NewBlock(idInfo, name, typeInfo)}}}
	c.Init(c)
	return c
}

func (c *Cobweb) Clone() Behavior {
	cl := *c
	cl.rebind(&cl)
	return &cl
}

func (c *Cobweb) HasEntityCollision() bool { return true }

func (c *Cobweb) OnEntityInside(entity Entity) bool {
	entity.ResetFallDistance()
	return true
}

// GetDropsForCompatibleTool is a port of Cobweb::getDropsForCompatibleTool. The non-shears branch
// (VanillaItems.STRING(), a pure item with no corresponding block) can't be built from here -
// there's no NewItemBlockFunc-style factory for standalone items, only for ItemBlock - so that
// branch stays a documented gap; only the shears branch (the block's own AsItem form) is real.
func (c *Cobweb) GetDropsForCompatibleTool(item Item) []Item {
	if item.GetBlockToolType()&ToolTypeShears == 0 {
		return nil
	}
	dropped := asItemOrNil(c.self)
	if dropped == nil {
		return nil
	}
	return []Item{dropped}
}

func (c *Cobweb) IsAffectedBySilkTouch() bool { return true }

func (c *Cobweb) BlocksDirectSkyLight() bool { return true }
