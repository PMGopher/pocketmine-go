package block

import (
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
)

// Loom is a port of pocketmine\block\Loom.
type Loom struct {
	Opaque
	HorizontalFacingComponent
}

func NewLoom(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Loom {
	l := &Loom{
		Opaque:                    Opaque{NewBlock(idInfo, name, typeInfo)},
		HorizontalFacingComponent: NewHorizontalFacingComponent(),
	}
	l.Init(l)
	return l
}

func (l *Loom) Clone() Behavior {
	c := *l
	c.rebind(&c)
	return &c
}

func (l *Loom) DescribeBlockOnlyState(w runtime.DataDescriber) { l.DescribeHorizontalFacing(w) }

// Place is a port of pocketmine\block\utils\FacesOppositePlacingPlayerTrait::place.
func (l *Loom) Place(tx BlockTransaction, item Item, blockReplace Behavior, blockClicked Behavior, face math.Facing, clickVector math.Vector3, player Player) bool {
	if player != nil {
		l.Facing = math.Opposite(player.GetHorizontalFacing())
	}
	return l.Block.Place(tx, item, blockReplace, blockClicked, face, clickVector, player)
}

// OnInteract should open a LoomInventory for the interacting player — needs the unported
// block/inventory package, so this is a no-op for now; it still returns whether a player was
// present, matching the PHP original.
func (l *Loom) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	return player != nil
}
