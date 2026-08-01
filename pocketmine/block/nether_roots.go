package block

import "pocketmine-go/pocketmine/math"

// NetherRoots is a port of pocketmine\block\NetherRoots.
type NetherRoots struct {
	Flowable
}

func NewNetherRoots(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *NetherRoots {
	n := &NetherRoots{Flowable{Transparent{NewBlock(idInfo, name, typeInfo)}}}
	n.Init(n)
	return n
}

func (n *NetherRoots) Clone() Behavior {
	c := *n
	c.rebind(&c)
	return &c
}

func (n *NetherRoots) canBeSupportedAt(blk Behavior) bool {
	// TODO: moss
	support := blk.(blockGeometry).GetSide(math.Down, 1)
	geo := support.(blockGeometry)
	return geo.HasTypeTag(BlockTypeTagsDirt) || geo.HasTypeTag(BlockTypeTagsMud) ||
		geo.HasTypeTag(BlockTypeTagsNylium) || support.GetTypeId() == SOUL_SOIL
}

func (n *NetherRoots) CanBePlacedAt(blockReplace Behavior, clickVector math.Vector3, face math.Facing, isClickedBlock bool) bool {
	return n.canBeSupportedAt(blockReplace) && n.Flowable.CanBePlacedAt(blockReplace, clickVector, face, isClickedBlock)
}

func (n *NetherRoots) OnNearbyBlockChange() {
	if !n.canBeSupportedAt(n.self) {
		if world, err := n.position.GetWorld(); err == nil {
			world.UseBreakOn(n.position.AsVector3())
		}
	} else {
		n.Flowable.OnNearbyBlockChange()
	}
}
