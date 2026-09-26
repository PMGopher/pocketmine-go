package block

import "pocketmine-go/pocketmine/math"

// NetherFungus is a port of pocketmine\block\NetherFungus.
type NetherFungus struct {
	Flowable

	// TreeType is the huge fungus it grows into.
	TreeType TreeType
	// NyliumTypeID is the type id of the nylium block this fungus can grow on.
	NyliumTypeID int
}

func NewNetherFungus(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo, treeType TreeType, nyliumTypeID int) *NetherFungus {
	n := &NetherFungus{Flowable: Flowable{Transparent{NewBlock(idInfo, name, typeInfo)}}, TreeType: treeType, NyliumTypeID: nyliumTypeID}
	n.Init(n)
	return n
}

func (n *NetherFungus) Clone() Behavior {
	c := *n
	c.rebind(&c)
	return &c
}

func (n *NetherFungus) canBeSupportedAt(blk Behavior) bool {
	// TODO: moss
	support := blk.(blockGeometry).GetSide(math.Down, 1)
	geo := support.(blockGeometry)
	return geo.HasTypeTag(BlockTypeTagsDirt) || geo.HasTypeTag(BlockTypeTagsMud) ||
		geo.HasTypeTag(BlockTypeTagsNylium) || support.GetTypeId() == SOUL_SOIL
}

func (n *NetherFungus) CanBePlacedAt(blockReplace Behavior, clickVector math.Vector3, face math.Facing, isClickedBlock bool) bool {
	return n.canBeSupportedAt(blockReplace) && n.Flowable.CanBePlacedAt(blockReplace, clickVector, face, isClickedBlock)
}

func (n *NetherFungus) OnNearbyBlockChange() {
	if !n.canBeSupportedAt(n.self) {
		if world, err := n.position.GetWorld(); err == nil {
			world.UseBreakOn(n.position.AsVector3())
		}
	} else {
		n.Flowable.OnNearbyBlockChange()
	}
}

// OnInteract is a port of NetherFungus::onInteract: bone meal has a 40% chance (always in
// creative) to grow a huge fungus if it's on its nylium.
func (n *NetherFungus) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	if isFertilizer(item) {
		item.Pop()
		if n.self.(blockGeometry).GetSide(math.Down, 1).GetTypeId() == n.NyliumTypeID && (player == nil || !hasFiniteResources(player) || mtRand(1, 100) <= 40) {
			growStructure(n.self, n.TreeType, player)
		}
		return true
	}
	return false
}
