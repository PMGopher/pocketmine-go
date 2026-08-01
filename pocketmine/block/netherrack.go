package block

import "pocketmine-go/pocketmine/math"

// Netherrack is a port of pocketmine\block\Netherrack.
type Netherrack struct {
	Opaque
}

func NewNetherrack(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *Netherrack {
	n := &Netherrack{Opaque{NewBlock(idInfo, name, typeInfo)}}
	n.Init(n)
	return n
}

func (n *Netherrack) Clone() Behavior {
	c := *n
	c.rebind(&c)
	return &c
}

func (n *Netherrack) BurnsForever() bool { return true }

// OnInteract should fertilizer-transform this into Warped/Crimson Nylium when a nearby nylium
// block is found (tryTransform) - needs a Fertilizer item marker, World.GetBlock(Position), and
// the block registry (VanillaBlocks), none ported yet, so this is a no-op for now.
func (n *Netherrack) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	return false
}
