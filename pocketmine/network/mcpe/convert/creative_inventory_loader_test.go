package convert

import (
	"testing"

	"pocketmine-go/pocketmine/inventory"
	"pocketmine-go/pocketmine/item"
)

func TestCreativeInventoryHasTheSupportedVanillaItems(t *testing.T) {
	inv := inventory.NewCreativeInventory()
	loadCreativeInventory(inv)
	if !inv.Contains(item.VanillaApple()) {
		t.Error("the creative inventory doesn't contain an apple")
	}
	entries, _ := inv.GetAllEntries()
	if len(entries) < 50 {
		t.Errorf("only %d creative items loaded", len(entries))
	}
}

func TestNetItemStackRoundTrip(t *testing.T) {
	stick := item.VanillaStick()
	stick.SetCount(12)
	back, err := NetItemStackToCore(CoreItemStackToNet(stick))
	if err != nil {
		t.Fatal(err)
	}
	if !back.EqualsExact(stick) {
		t.Errorf("round trip: got %v, want %v", back, stick)
	}
}
