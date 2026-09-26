package io

import (
	"testing"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/nbt"
)

func roundTrip(t *testing.T, it item.Item, slot int) item.Item {
	t.Helper()
	tag, err := item.NbtSerialize(it, slot)
	if err != nil {
		t.Fatalf("serializing %s: %v", it.GetName(), err)
	}
	got, err := item.NbtDeserialize(tag)
	if err != nil {
		t.Fatalf("deserializing %s (%s): %v", it.GetName(), tag, err)
	}
	return got
}

func TestItemNbtRoundTrip(t *testing.T) {
	diamond := item.VanillaItem("diamond")
	diamond.SetCount(5)
	sword := item.VanillaItem("diamond_sword")
	sword.(interface{ SetDamage(int) }).SetDamage(17)
	sword.SetCustomName("Excalibur")
	planks, err := block.VanillaBlock("birch_planks").(interface{ AsItem() (block.Item, error) }).AsItem()
	if err != nil {
		t.Fatal(err)
	}
	for _, it := range []item.Item{diamond, sword, planks.(item.Item)} {
		got := roundTrip(t, it, 3)
		if !got.EqualsExact(it) {
			t.Errorf("%s (count %d) came back as %s (count %d)", it.GetName(), it.GetCount(), got.GetName(), got.GetCount())
		}
	}

	tag, _ := item.NbtSerialize(diamond, 3)
	if slot, err := tag.GetByte("Slot"); err != nil || slot != 3 {
		t.Errorf("Slot tag = %v (%v)", slot, err)
	}
	if name, _ := tag.GetString("Name"); name != "minecraft:diamond" {
		t.Errorf("Name tag = %s", name)
	}
	tag, _ = item.NbtSerialize(diamond, -1)
	if _, ok := tag.GetTag("Slot"); ok {
		t.Errorf("a stack without a slot shouldn't have a Slot tag")
	}
}

func TestItemNbtDeserializesLegacyFormats(t *testing.T) {
	// Bedrock <= 1.5 / PocketMine-MP 3: numeric ID + Damage.
	legacySword := nbt.NewCompoundTag()
	legacySword.SetShort("id", 276)
	legacySword.SetShort("Damage", 12)
	legacySword.SetByte("Count", 1)
	got, err := item.NbtDeserialize(legacySword)
	if err != nil {
		t.Fatal(err)
	}
	if got.GetName() != "Diamond Sword" || got.(interface{ GetDamage() int }).GetDamage() != 12 {
		t.Errorf("legacy 276:12 came back as %s", got.GetName())
	}

	// Legacy block item: planks:2 is birch planks.
	legacyPlanks := nbt.NewCompoundTag()
	legacyPlanks.SetShort("id", 5)
	legacyPlanks.SetShort("Damage", 2)
	legacyPlanks.SetByte("Count", 64)
	got, err = item.NbtDeserialize(legacyPlanks)
	if err != nil {
		t.Fatal(err)
	}
	if got.GetName() != "Birch Planks" || got.GetCount() != 64 {
		t.Errorf("legacy 5:2 x64 came back as %s x%d", got.GetName(), got.GetCount())
	}

	// id 0 (air saved by old versions) is air, not an error.
	air := nbt.NewCompoundTag()
	air.SetShort("id", 0)
	air.SetByte("Count", 0)
	if got, err := item.NbtDeserialize(air); err != nil || !got.IsNull() {
		t.Errorf("legacy air: %v, %v", got, err)
	}

	// Unknown items are errors, which SafeNbtDeserialize replaces with air.
	unknown := nbt.NewCompoundTag()
	unknown.SetString("Name", "minecraft:not_an_item")
	unknown.SetByte("Count", 1)
	if _, err := item.NbtDeserialize(unknown); err == nil {
		t.Errorf("unknown item should fail")
	}
	if got := item.SafeNbtDeserialize(unknown, "test", nil); !got.IsNull() {
		t.Errorf("SafeNbtDeserialize of an unknown item returned %s", got.GetName())
	}
}
