package block

import "pocketmine-go/pocketmine/math"

// Item type IDs (pocketmine\item\ItemTypeIds) the cauldrons react to. This package can't import
// item (item imports block), so the values are copied from item/type_ids.go.
const (
	itemTypeIDsBucket           = 20023
	itemTypeIDsGlassBottle      = 20109
	itemTypeIDsLavaBucket       = 20142
	itemTypeIDsPotion           = 20166
	itemTypeIDsSplashPotion     = 20202
	itemTypeIDsWaterBucket      = 20218
	itemTypeIDsPowderSnowBucket = 20258
	itemTypeIDsLingeringPotion  = 20259
	itemTypeIDsBamboo           = 20005
)

// VanillaItemFunc returns a new instance of the named VanillaItems entry (e.g. "bucket"). This
// package can't import item, so the item package sets it in init(). Nil in tests that don't import
// item; then blocks that hand items back (cauldrons) return nothing.
var VanillaItemFunc func(name string) Item

func vanillaItem(name string) Item {
	if VanillaItemFunc == nil {
		return nil
	}
	return VanillaItemFunc(name)
}

// waterPotionChecker is the part of item.Potion/item.SplashPotion the cauldrons need
// ($item->getType() === PotionType::WATER).
type waterPotionChecker interface {
	IsWaterPotion() bool
}

// cauldronCollisionBoxes is Cauldron/FillableCauldron::recalculateCollisionBoxes.
func cauldronCollisionBoxes() []math.AxisAlignedBB {
	result := []math.AxisAlignedBB{
		math.OneAABB().TrimmedCopy(math.Up, 11.0/16), //bottom of the cauldron
	}
	for _, f := range math.HorizontalFacing { //add the frame parts around the bowl
		result = append(result, math.OneAABB().TrimmedCopy(f, 14.0/16))
	}
	return result
}

func appendItem(items *[]Item, it Item) {
	if items != nil && it != nil {
		*items = append(*items, it)
	}
}

// vanillaItemCount is VanillaItems::X()->setCount($count) (nil if the item package isn't loaded).
func vanillaItemCount(name string, count int) Item {
	it := vanillaItem(name)
	if it != nil {
		it.SetCount(count)
	}
	return it
}

// itemDrops is a one-item drop list, empty if the item couldn't be built.
func itemDrops(it Item) []Item {
	if it == nil {
		return nil
	}
	return []Item{it}
}
