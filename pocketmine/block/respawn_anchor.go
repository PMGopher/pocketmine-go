package block

import (
	runtime "pocketmine-go/pocketmine/data/runtime"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/world/sound"
)

const (
	respawnAnchorMinCharges = 0
	respawnAnchorMaxCharges = 4
)

// itemTypeIDsGlowstone mirrors ItemTypeIds::fromBlockTypeId(BlockTypeIds::GLOWSTONE), i.e.
// -GLOWSTONE (negative item type IDs are treated as block IDs in the PHP original) - not imported
// from a real item package since it doesn't exist yet, same reasoning as itemTypeIDsHoneycomb.
const itemTypeIDsGlowstone = -GLOWSTONE

// RespawnAnchor is a port of pocketmine\block\RespawnAnchor.
type RespawnAnchor struct {
	Opaque

	Charges int
}

func NewRespawnAnchor(idInfo *BlockIdentifier, name string, typeInfo *BlockTypeInfo) *RespawnAnchor {
	r := &RespawnAnchor{Opaque: Opaque{NewBlock(idInfo, name, typeInfo)}, Charges: respawnAnchorMinCharges}
	r.Init(r)
	return r
}

func (r *RespawnAnchor) Clone() Behavior {
	c := *r
	c.rebind(&c)
	return &c
}

func (r *RespawnAnchor) DescribeBlockOnlyState(w runtime.DataDescriber) {
	charges := r.Charges
	w.BoundedIntAuto(respawnAnchorMinCharges, respawnAnchorMaxCharges, &charges)
	r.Charges = charges
}

func (r *RespawnAnchor) GetCharges() int { return r.Charges }

func (r *RespawnAnchor) SetCharges(charges int) {
	if charges < respawnAnchorMinCharges || charges > respawnAnchorMaxCharges {
		panic("Charges must be between 0 and 4")
	}
	r.Charges = charges
}

func (r *RespawnAnchor) GetLightLevel() int {
	if r.Charges > 0 {
		return r.Charges*4 - 1
	}
	return 0
}

// OnInteract's charging branch (right-clicking with glowstone) is fully ported. The
// explode/set-spawn branch (right-clicking while charged) needs
// PlayerRespawnAnchorUseEvent/BlockPreExplodeEvent, the Explosion subsystem, and
// Player.GetSpawn/SetSpawn, none ported yet, so it's a documented no-op for now.
func (r *RespawnAnchor) OnInteract(item Item, face math.Facing, clickVector math.Vector3, player Player, returnedItems *[]Item) bool {
	if item.GetTypeId() == itemTypeIDsGlowstone && r.Charges < respawnAnchorMaxCharges {
		world, err := r.position.GetWorld()
		if err != nil {
			return false
		}
		r.Charges++
		if err := world.SetBlock(r.position, r.self); err != nil {
			panic(err)
		}
		world.AddSound(r.position.AsVector3(), sound.RespawnAnchorChargeSound{})
		item.Pop()
		return true
	}
	return false
}
