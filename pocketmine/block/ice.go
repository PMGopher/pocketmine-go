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

// OnBreak is a port of Ice::onBreak. Item.HasEnchantment(SILK_TOUCH) isn't ported (the
// enchantment system isn't ported at all yet - same "always false" convention as every other
// HasEnchantment check in this port, see candle_component.go), so the silk-touch exemption never
// applies; the survival-only gate is otherwise real.
func (i *Ice) OnBreak(item Item, player Player, returnedItems *[]Item) bool {
	if player == nil || player.IsSurvival() {
		Melt(i.self, VanillaWater())
		return true
	}
	return i.Block.OnBreak(item, player, returnedItems)
}

func (i *Ice) TicksRandomly() bool { return true }

// OnRandomTick is a port of Ice::onRandomTick.
func (i *Ice) OnRandomTick() {
	world, err := i.position.GetWorld()
	if err != nil {
		return
	}
	pos := i.position.AsVector3()
	if world.GetHighestAdjacentBlockLightAt(pos.FloorX(), pos.FloorY(), pos.FloorZ()) >= 12 {
		Melt(i.self, VanillaWater())
	}
}

// GetDropsForCompatibleTool deliberately returns nothing — ice melts instead of dropping itself,
// matching the PHP original's `return [];` (this isn't a not-yet-ported gap).
func (i *Ice) GetDropsForCompatibleTool(item Item) []Item { return nil }

func (i *Ice) IsAffectedBySilkTouch() bool { return true }
