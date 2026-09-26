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

// OnInteract is a port of CraftingTable::onInteract.
func (c *CraftingTable) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	openWindow(player, WindowCraftingTable, c.position)
	return true
}

func (c *CraftingTable) GetFuelTime() int { return 300 }
