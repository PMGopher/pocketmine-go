package transaction

import (
	"testing"

	"pocketmine-go/pocketmine/inventory"
	"pocketmine-go/pocketmine/item"
	"pocketmine-go/pocketmine/math"
)

type fakePlayer struct {
	finite  bool
	dropped []item.Item
}

func (p *fakePlayer) GetID() int                { return 1 }
func (p *fakePlayer) GetPosition() math.Vector3 { return math.Vector3{} }
func (p *fakePlayer) IsClosed() bool            { return false }
func (p *fakePlayer) GetName() string           { return "Steve" }
func (p *fakePlayer) GetDisplayName() string    { return "Steve" }
func (p *fakePlayer) HasFiniteResources() bool  { return p.finite }
func (p *fakePlayer) GetCreativeInventory() *inventory.CreativeInventory {
	return inventory.NewCreativeInventory()
}
func (p *fakePlayer) IsSpectator() bool     { return false }
func (p *fakePlayer) DropItem(it item.Item) { p.dropped = append(p.dropped, it) }

func stick(count int) item.Item {
	s := item.VanillaStick()
	s.SetCount(count)
	return s
}

func TestMoveBetweenInventoriesBalances(t *testing.T) {
	a, b := inventory.NewSimpleInventory(2), inventory.NewSimpleInventory(2)
	a.SetItem(0, stick(10))

	builder := NewTransactionBuilder()
	builder.GetInventory(a).SetItem(0, stick(4))
	builder.GetInventory(b).SetItem(1, stick(6))
	tx, err := NewInventoryTransaction(&fakePlayer{finite: true}, builder.GenerateActions())
	if err != nil {
		t.Fatal(err)
	}
	if err := tx.Execute(); err != nil {
		t.Fatal(err)
	}
	if a.GetItem(0).GetCount() != 4 || b.GetItem(1).GetCount() != 6 {
		t.Fatalf("a=%d b=%d", a.GetItem(0).GetCount(), b.GetItem(1).GetCount())
	}
}

func TestUnbalancedTransactionIsRejected(t *testing.T) {
	a := inventory.NewSimpleInventory(1)
	tx, _ := NewInventoryTransaction(&fakePlayer{finite: true}, []InventoryAction{NewSlotChangeAction(a, 0, inventory.Air(), stick(5))})
	err := tx.Execute()
	if _, ok := err.(*TransactionValidationError); !ok {
		t.Fatalf("expected a validation error, got %v", err)
	}
	if !a.IsSlotEmpty(0) {
		t.Fatalf("a rejected transaction must not change anything")
	}
}

func TestDropAndSquash(t *testing.T) {
	a := inventory.NewSimpleInventory(1)
	a.SetItem(0, stick(3))
	p := &fakePlayer{finite: true}
	tx, _ := NewInventoryTransaction(p, []InventoryAction{
		NewSlotChangeAction(a, 0, stick(3), stick(2)),
		NewSlotChangeAction(a, 0, stick(2), inventory.Air()),
		NewDropItemAction(stick(3)),
	})
	if err := tx.Execute(); err != nil {
		t.Fatal(err)
	}
	if !a.IsSlotEmpty(0) || len(p.dropped) != 1 || p.dropped[0].GetCount() != 3 {
		t.Fatalf("slot empty=%v dropped=%v", a.IsSlotEmpty(0), p.dropped)
	}
}
