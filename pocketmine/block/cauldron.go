package block

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/math"
)

// Cauldron is a port of pocketmine\block\Cauldron (the empty cauldron).
//
// writeStateToWorld's tile reset (TileCauldron::setCustomWaterColor/setPotionItem) isn't ported:
// there is no Cauldron tile yet.
type Cauldron struct {
	Transparent
}

func NewCauldron(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Cauldron {
	c := &Cauldron{Transparent{NewBlock(idInfo, name, typeInfo)}}
	c.Init(c)
	return c
}

func (c *Cauldron) Clone() Behavior {
	cp := *c
	cp.rebind(&cp)
	return &cp
}

func (c *Cauldron) RecalculateCollisionBoxes() []math.AxisAlignedBB { return cauldronCollisionBoxes() }

func (c *Cauldron) GetSupportType(facing math.Facing) blockutils.SupportType {
	if facing == math.Up {
		return blockutils.SupportTypeEdge
	}
	return blockutils.SupportTypeNone
}

// fill is a port of Cauldron::fill.
func (c *Cauldron) fill(amount int, result fillableCauldron, usedItem Item, returnedItem Item, returnedItems *[]Item) {
	world, err := c.position.GetWorld()
	if err != nil {
		return
	}
	result.SetFillLevel(amount)
	_ = world.SetBlock(c.position, result)
	world.AddSound(c.position.Add(0.5, 0.5, 0.5), result.GetFillSound())

	usedItem.Pop()
	appendItem(returnedItems, returnedItem)
}

// OnInteract is a port of Cauldron::onInteract.
func (c *Cauldron) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	switch item.GetTypeId() {
	case itemTypeIDsWaterBucket:
		c.fill(FillableCauldronMaxFillLevel, VanillaBlock("water_cauldron").(*WaterCauldron), item, vanillaItem("bucket"), returnedItems)
	case itemTypeIDsLavaBucket:
		c.fill(FillableCauldronMaxFillLevel, VanillaBlock("lava_cauldron").(*LavaCauldron), item, vanillaItem("bucket"), returnedItems)
	case itemTypeIDsPowderSnowBucket:
		//TODO: powder snow cauldron
	default:
		if potion, ok := item.(waterPotionChecker); ok {
			if potion.IsWaterPotion() {
				c.fill(WaterCauldronWaterBottleFillAmount, VanillaBlock("water_cauldron").(*WaterCauldron), item, vanillaItem("glass_bottle"), returnedItems)
			} else {
				potionCauldron := VanillaBlock("potion_cauldron").(*PotionCauldron)
				potionCauldron.SetPotionItem(item)
				c.fill(PotionCauldronPotionFillAmount, potionCauldron, item, vanillaItem("glass_bottle"), returnedItems)
			}
		}
	}
	return true
}

// OnNearbyBlockChange is a port of Cauldron::onNearbyBlockChange: water above fills the cauldron.
func (c *Cauldron) OnNearbyBlockChange() {
	world, err := c.position.GetWorld()
	if err != nil {
		return
	}
	if c.GetSide(math.Up, 1).GetTypeId() == WATER {
		cauldron := VanillaBlock("water_cauldron").(*WaterCauldron)
		cauldron.SetFillLevel(FillableCauldronMaxFillLevel)
		_ = world.SetBlock(c.position, cauldron)
		world.AddSound(c.position.Add(0.5, 0.5, 0.5), cauldron.GetFillSound())
	}
}
