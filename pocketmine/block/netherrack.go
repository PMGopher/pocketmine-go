package block

import (
	"math/rand"

	"pocketmine-go/pocketmine/math"
)

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

// OnInteract is a port of Netherrack::onInteract. `$item instanceof Fertilizer` is checked via
// item type ID (bone meal is the only Fertilizer-marked item in the PHP original), same
// structural-marker convention as Crops.OnInteract.
func (n *Netherrack) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	if item.GetTypeId() == itemTypeIDsBoneMeal && n.tryTransform() {
		item.Pop()
		return true
	}
	return false
}

// tryTransform is a port of Netherrack::tryTransform.
func (n *Netherrack) tryTransform() bool {
	world, err := n.position.GetWorld()
	if err != nil {
		return false
	}
	pos := n.position

	if !world.GetBlockAt(pos.FloorX(), pos.FloorY()+1, pos.FloorZ()).IsTransparent() {
		return false
	}

	hasWarpedNylium, hasCrimsonNylium := false, false
outer:
	for x := -1; x <= 1; x++ {
		for y := -1; y <= 1; y++ {
			for z := -1; z <= 1; z++ {
				switch world.GetBlockAt(pos.FloorX()+x, pos.FloorY()+y, pos.FloorZ()+z).GetTypeId() {
				case WARPED_NYLIUM:
					hasWarpedNylium = true
				case CRIMSON_NYLIUM:
					hasCrimsonNylium = true
				}
				if hasWarpedNylium && hasCrimsonNylium {
					break outer
				}
			}
		}
	}

	if !hasWarpedNylium && !hasCrimsonNylium {
		return false
	}

	var newState Behavior
	switch {
	case hasWarpedNylium && hasCrimsonNylium:
		if rand.Intn(2) == 0 {
			newState = VanillaWarpedNylium()
		} else {
			newState = VanillaCrimsonNylium()
		}
	case hasWarpedNylium:
		newState = VanillaWarpedNylium()
	default:
		newState = VanillaCrimsonNylium()
	}
	_ = world.SetBlock(n.position, newState)
	return true
}
