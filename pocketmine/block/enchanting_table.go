package block

import (
	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/math"
)

// EnchantingTable is a port of pocketmine\block\EnchantingTable, minus actually opening the
// inventory window (player.SetCurrentWindow isn't ported - see block.Chest.OnInteract's doc
// comment for the same gap) and the "//TODO lock" the PHP original never implemented either.
type EnchantingTable struct {
	Transparent
}

func NewEnchantingTable(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *EnchantingTable {
	e := &EnchantingTable{Transparent: Transparent{NewBlock(idInfo, name, typeInfo)}}
	e.Init(e)
	return e
}

func (e *EnchantingTable) Clone() Behavior {
	c := *e
	c.rebind(&c)
	return &c
}

func (e *EnchantingTable) RecalculateCollisionBoxes() []math.AxisAlignedBB {
	return []math.AxisAlignedBB{math.OneAABB().TrimmedCopy(math.Up, 0.25)}
}

func (e *EnchantingTable) GetSupportType(facing math.Facing) blockutils.SupportType {
	return blockutils.SupportTypeNone
}

func (e *EnchantingTable) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	// player.SetCurrentWindow(NewEnchantInventory(e.position)) - not ported, see doc comment above.
	return true
}
