package block

import (
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/world/sound"
)

const PotionCauldronPotionFillAmount = 2

// PotionCauldron is a port of pocketmine\block\PotionCauldron. The potion item is kept on the
// block; saving it in the Cauldron tile (readStateFromWorld/writeStateToWorld) isn't ported (no
// Cauldron tile yet).
type PotionCauldron struct {
	FillableCauldron

	potionItem Item
}

func NewPotionCauldron(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *PotionCauldron {
	p := &PotionCauldron{FillableCauldron: newFillableCauldron(idInfo, name, typeInfo)}
	p.Init(p)
	return p
}

func (p *PotionCauldron) Clone() Behavior {
	c := *p
	c.rebind(&c)
	return &c
}

// GetPotionItem is a port of PotionCauldron::getPotionItem.
func (p *PotionCauldron) GetPotionItem() Item { return p.potionItem }

// SetPotionItem is a port of PotionCauldron::setPotionItem.
func (p *PotionCauldron) SetPotionItem(potionItem Item) {
	if potionItem != nil {
		switch potionItem.GetTypeId() {
		case itemTypeIDsPotion, itemTypeIDsSplashPotion, itemTypeIDsLingeringPotion:
		default:
			panic("Item must be a POTION, SPLASH_POTION or LINGERING_POTION")
		}
	}
	p.potionItem = potionItem
}

func (p *PotionCauldron) GetFillSound() sound.Sound { return sound.CauldronFillPotionSound{} }

func (p *PotionCauldron) GetEmptySound() sound.Sound { return sound.CauldronEmptyPotionSound{} }

// OnInteract is a port of PotionCauldron::onInteract. addFillLevelsOrMix's
// $usedItem->equals($this->potionItem) check compares type IDs only here (the block package can't
// compare item states).
func (p *PotionCauldron) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	switch item.GetTypeId() {
	case itemTypeIDsLingeringPotion, itemTypeIDsPotion, itemTypeIDsSplashPotion:
		if p.potionItem != nil && item.GetTypeId() != p.potionItem.GetTypeId() {
			p.mix(item, vanillaItem("glass_bottle"), returnedItems)
		} else {
			p.addFillLevels(PotionCauldronPotionFillAmount, item, vanillaItem("glass_bottle"), returnedItems)
		}
	case itemTypeIDsGlassBottle:
		if p.potionItem != nil {
			p.removeFillLevels(PotionCauldronPotionFillAmount, item, p.potionItem, returnedItems)
		}
	case itemTypeIDsLavaBucket, itemTypeIDsPowderSnowBucket, itemTypeIDsWaterBucket:
		p.mix(item, vanillaItem("bucket"), returnedItems)
		//TODO: tipped arrows
	}
	return true
}

// OnNearbyBlockChange is a port of PotionCauldron::onNearbyBlockChange.
func (p *PotionCauldron) OnNearbyBlockChange() {
	world, err := p.position.GetWorld()
	if err != nil {
		return
	}
	if p.GetSide(math.Up, 1).GetTypeId() == WATER {
		cauldron := VanillaBlock("water_cauldron").(*WaterCauldron)
		cauldron.SetFillLevel(FillableCauldronMaxFillLevel)
		_ = world.SetBlock(p.position, cauldron)
		world.AddSound(p.position.Add(0.5, 0.5, 0.5), cauldron.GetFillSound())
	}
}
