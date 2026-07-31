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

// GetDropsForCompatibleTool should return [c.AsItem()] for shears and [item.VanillaItems.String()]
// otherwise — both need real Item construction from the unported item package (see
// Block.GetDropsForCompatibleTool's doc comment), so this returns nil for now.
func (c *Cobweb) GetDropsForCompatibleTool(item Item) []Item { return nil }

func (c *Cobweb) IsAffectedBySilkTouch() bool { return true }

func (c *Cobweb) BlocksDirectSkyLight() bool { return true }
