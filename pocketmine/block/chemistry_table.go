package block

import (
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// ChemistryTable is a port of pocketmine\block\ChemistryTable.
type ChemistryTable struct {
	Opaque
	HorizontalFacingComponent
}

func NewChemistryTable(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *ChemistryTable {
	c := &ChemistryTable{
		Opaque:                    Opaque{NewBlock(idInfo, name, typeInfo)},
		HorizontalFacingComponent: NewHorizontalFacingComponent(),
	}
	c.Init(c)
	return c
}

func (c *ChemistryTable) Clone() Behavior {
	cl := *c
	cl.rebind(&cl)
	return &cl
}

func (c *ChemistryTable) DescribeBlockOnlyState(w runtime.DataDescriber) {
	c.DescribeHorizontalFacing(w)
}

// Place is a port of pocketmine\block\utils\FacesOppositePlacingPlayerTrait::place.
func (c *ChemistryTable) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	if player != nil {
		c.Facing = math.Opposite(player.GetHorizontalFacing())
	}
	return c.Block.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
}

// OnInteract is a port of ChemistryTable::onInteract - unimplemented in the PHP original too
// (marked //TODO), so this stays a faithful always-false stub.
func (c *ChemistryTable) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	return false
}
