package block

import "pocketmine-go/pocketmine/math"

// NetherFungus is a port of pocketmine\block\NetherFungus.
//
// The PHP constructor's $treeType param is omitted here: it's only ever used by grow() (via
// TreeFactory), which needs the unported world-gen tree subsystem and is a documented no-op
// below, so there's nothing to store it for yet.
type NetherFungus struct {
	Flowable

	// NyliumTypeID is the type id of the nylium block this fungus can grow on.
	NyliumTypeID int
}

func NewNetherFungus(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo, nyliumTypeID int) *NetherFungus {
	n := &NetherFungus{Flowable: Flowable{Transparent{NewBlock(idInfo, name, typeInfo)}}, NyliumTypeID: nyliumTypeID}
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

// OnInteract's fertilizer-driven grow needs a Fertilizer item marker, StructureGrowEvent, and the
// world-gen tree subsystem (TreeFactory/TreeType), none ported yet. Block's default OnInteract
// (return false) already matches this gap, so there's nothing to override here.
