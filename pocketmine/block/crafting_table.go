package block

import "pocketmine-go/pocketmine/math"

// CraftingTable is a port of pocketmine\block\CraftingTable.
type CraftingTable struct {
	Opaque
}

func NewCraftingTable(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *CraftingTable {
	c := &CraftingTable{Opaque{NewBlock(idInfo, name, typeInfo)}}
	c.Init(c)
	return c
}

func (c *CraftingTable) Clone() Behavior {
	cl := *c
	cl.rebind(&cl)
	return &cl
}

// OnInteract should open a CraftingTableInventory for the interacting player — needs the unported
// block/inventory package, so this is a no-op for now; it still returns true, matching the PHP
// original's unconditional `return true;`.
func (c *CraftingTable) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	return true
}

func (c *CraftingTable) GetFuelTime() int { return 300 }
