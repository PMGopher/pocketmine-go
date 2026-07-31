package block

// Ice is a port of pocketmine\block\Ice.
type Ice struct {
	Transparent
}

func NewIce(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Ice {
	i := &Ice{Transparent{NewBlock(idInfo, name, typeInfo)}}
	i.Init(i)
	return i
}

func (i *Ice) Clone() Behavior {
	c := *i
	c.rebind(&c)
	return &c
}

func (i *Ice) GetLightFilter() int { return 2 }

func (i *Ice) GetFrictionFactor() float64 { return 0.98 }

// OnBreak should replace itself with VanillaBlocks.WATER() when broken without silk touch by a
// survival (or console) player — needs Item.HasEnchantment and the not-yet-ported block registry
// (VanillaBlocks), so this always falls back to the default (no water left behind) for now.
func (i *Ice) OnBreak(item Item, player Player, returnedItems *[]Item) bool {
	return i.Block.OnBreak(item, player, returnedItems)
}

func (i *Ice) TicksRandomly() bool { return true }

// OnRandomTick should melt into VanillaBlocks.WATER() via BlockEventHelper when adjacent light is
// high enough — needs World.GetHighestAdjacentBlockLight and the block registry, neither ported
// yet, so this is a no-op for now (same category of gap as Vine's growth TODO).
func (i *Ice) OnRandomTick() {}

// GetDropsForCompatibleTool deliberately returns nothing — ice melts instead of dropping itself,
// matching the PHP original's `return [];` (this isn't a not-yet-ported gap).
func (i *Ice) GetDropsForCompatibleTool(item Item) []Item { return nil }

func (i *Ice) IsAffectedBySilkTouch() bool { return true }
