package player

import (
	"testing"

	"pocketmine-go/pocketmine/block"
	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/math"
)

func TestWaterCauldronDyesAndCleansLeatherArmor(t *testing.T) {
	p := newTestPlayer(t, 1, math.NewVector3(0.5, 70, 0.5))
	setBlockAt(t, p, 0, 70, 2, block.VanillaBlock("water_cauldron").(*block.WaterCauldron).SetCustomWaterColor(nil))
	cauldron := func() *block.WaterCauldron {
		c, ok := p.GetWorld().GetBlockAt(0, 70, 2).(*block.WaterCauldron)
		if !ok {
			t.Fatalf("block = %v, want a water cauldron", p.GetWorld().GetBlockAt(0, 70, 2))
		}
		return c
	}
	use := func(it item.Item) []block.Item {
		var returned []block.Item
		cauldron().OnInteract(it, math.Up, math.NewVector3(0.5, 1, 0.5), p, &returned)
		return returned
	}
	c := cauldron()
	c.SetFillLevel(block.FillableCauldronMaxFillLevel)
	setBlockAt(t, p, 0, 70, 2, c)

	dye := item.VanillaItem("dye")
	dye.(*item.Dye).SetColor(blockutils.DyeColorRed)
	use(dye)
	waterColor := cauldron().GetCustomWaterColor()
	if waterColor == nil || !waterColor.Equals(blockutils.DyeColorRed.GetRgbValue()) {
		t.Fatalf("water colour = %v, want red", waterColor)
	}
	if dye.GetCount() != 0 {
		t.Error("the dye wasn't used up")
	}

	tunic := item.VanillaItem("leather_tunic").(*item.Armor)
	use(tunic)
	if got, ok := tunic.GetCustomColor(); !ok || !got.Equals(*waterColor) {
		t.Errorf("tunic colour = %v, want the water's red", got)
	}
	if got := cauldron().FillLevel; got != block.FillableCauldronMaxFillLevel-block.WaterCauldronDyeArmorUseAmount {
		t.Errorf("fill level after dyeing = %d", got)
	}

	// A water bucket clears the colour; clear water then washes the dye off.
	use(item.VanillaItem("water_bucket"))
	if cauldron().GetCustomWaterColor() != nil {
		t.Fatal("a water bucket didn't clear the water colour")
	}
	use(tunic)
	if _, ok := tunic.GetCustomColor(); ok {
		t.Error("clear water didn't wash the tunic")
	}

	// Glass bottles take water out as water bottles (PHP: POTION()->setType(WATER)).
	returned := use(item.VanillaGlassBottle())
	if len(returned) != 1 {
		t.Fatalf("glass bottle returned %v", returned)
	}
	if potion, ok := returned[0].(*item.Potion); !ok || potion.PotionTypeValue != item.PotionTypeWater {
		t.Errorf("returned %v, want a water bottle", returned[0])
	}
}
