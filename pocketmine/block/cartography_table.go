package block

import "pocketmine-go/pocketmine/math"

// CartographyTable is a port of pocketmine\block\CartographyTable, minus actually opening the
// inventory window (player.SetCurrentWindow isn't ported - see block.Chest.OnInteract's doc
// comment for the same gap).
type CartographyTable struct {
	Opaque
}

func NewCartographyTable(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *CartographyTable {
	c := &CartographyTable{Opaque: Opaque{NewBlock(idInfo, name, typeInfo)}}
	c.Init(c)
	return c
}

func (c *CartographyTable) Clone() Behavior {
	cl := *c
	cl.rebind(&cl)
	return &cl
}

func (c *CartographyTable) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	// player.SetCurrentWindow(NewCartographyTableInventory(c.position)) - not ported, see doc
	// comment above.
	return true
}

func (c *CartographyTable) GetFuelTime() int { return 300 }
