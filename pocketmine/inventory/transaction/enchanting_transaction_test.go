package transaction

import (
	"strings"
	"testing"

	"pocketmine-go/pocketmine/entity"
	"pocketmine-go/pocketmine/inventory"
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/item/enchantment"
)

type fakeEnchanter struct {
	fakePlayer
	seedRegenerated bool
}

func (p *fakeEnchanter) GetXpManager() *entity.ExperienceManager { return nil }
func (p *fakeEnchanter) RegenerateEnchantmentSeed()              { p.seedRegenerated = true }

func enchantingActions(inv *inventory.SimpleInventory, output item.Item, lapis int) []InventoryAction {
	inv.SetItem(0, item.VanillaBook())
	inv.SetItem(1, withCount(item.VanillaLapisLazuli(), 3))
	return []InventoryAction{
		NewSlotChangeAction(inv, 0, item.VanillaBook(), output),
		NewSlotChangeAction(inv, 1, withCount(item.VanillaLapisLazuli(), 3), withCount(item.VanillaLapisLazuli(), 3-lapis)),
	}
}

func TestEnchantingTransactionCreative(t *testing.T) {
	option := enchantment.NewEnchantingOption(1, "abc", []*enchantment.EnchantmentInstance{enchantment.NewEnchantmentInstance(enchantment.VanillaUnbreaking(), 1)})
	output := item.EnchantItem(item.VanillaBook(), option.GetEnchantments())
	inv := inventory.NewSimpleInventory(2)

	source := &fakeEnchanter{}
	tx := NewEnchantingTransaction(source, option, 1)
	for _, a := range enchantingActions(inv, output, 1) {
		if err := tx.AddAction(a); err != nil {
			t.Fatal(err)
		}
	}
	if err := tx.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if inv.GetItem(0).GetTypeId() != item.ENCHANTED_BOOK {
		t.Error("the book wasn't replaced by the enchanted book")
	}
	if !source.seedRegenerated {
		t.Error("the enchantment seed wasn't regenerated")
	}
}

func TestEnchantingTransactionRejectsWrongOutput(t *testing.T) {
	option := enchantment.NewEnchantingOption(1, "abc", []*enchantment.EnchantmentInstance{enchantment.NewEnchantmentInstance(enchantment.VanillaUnbreaking(), 1)})
	wrong := item.EnchantItem(item.VanillaBook(), []*enchantment.EnchantmentInstance{enchantment.NewEnchantmentInstance(enchantment.VanillaMending(), 1)})
	inv := inventory.NewSimpleInventory(2)

	tx := NewEnchantingTransaction(&fakeEnchanter{}, option, 1)
	for _, a := range enchantingActions(inv, wrong, 1) {
		if err := tx.AddAction(a); err != nil {
			t.Fatal(err)
		}
	}
	if err := tx.Execute(); err == nil || !strings.Contains(err.Error(), "Invalid output item") {
		t.Fatalf("Execute: got %v, want an invalid output error", err)
	}
}

func TestEnchantingTransactionChecksLapisInSurvival(t *testing.T) {
	option := enchantment.NewEnchantingOption(1, "abc", []*enchantment.EnchantmentInstance{enchantment.NewEnchantmentInstance(enchantment.VanillaUnbreaking(), 1)})
	output := item.EnchantItem(item.VanillaBook(), option.GetEnchantments())
	inv := inventory.NewSimpleInventory(2)

	tx := NewEnchantingTransaction(&fakeEnchanter{fakePlayer: fakePlayer{finite: true}}, option, 2)
	for _, a := range enchantingActions(inv, output, 1) {
		if err := tx.AddAction(a); err != nil {
			t.Fatal(err)
		}
	}
	if err := tx.Execute(); err == nil || !strings.Contains(err.Error(), "lapis lazuli spent to be 2, but received 1") {
		t.Fatalf("Execute: got %v, want a lapis error", err)
	}
}
