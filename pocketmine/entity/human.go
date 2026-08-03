package entity

import "pocketmine-go/pocketmine/math"

// humanEyeHeight mirrors Human::getInitialSizeInfo's own EntitySizeInfo(1.8, 0.6, 1.62) - a fixed
// eye height rather than the general EntitySizeInfo-derived one, since no other entity size exists
// in this port yet to make a general mechanism worthwhile (see Human's own doc comment).
const humanEyeHeight = 1.62

// Human is a port of a slice of pocketmine\entity\Human: real UUID, plus everything inherited from
// Living/Entity.
//
// Not ported (each is its own real subsystem with nowhere to plug into yet, matching this port's
// established "document the gap, don't guess" pattern - see the entity package's own doc comment
// for the identical reasoning applied to Entity/Living):
//   - HungerManager/ExperienceManager: no food/saturation ticking or XP orb absorption exists.
//   - inventory/offHandInventory/enderInventory: pocketmine/inventory (which needs pocketmine/item,
//     which needs pocketmine/block, which needs this package for DamageSource et al.) can't be
//     imported here without an import cycle - player.Player, one level up, holds its own real
//     inventory directly instead (a legitimate package-boundary adaptation, not a scope cut: the
//     inventory is just as real, one struct level higher than PHP's own class hierarchy puts it).
//   - Skin: network-only data (this port already handles skin encoding directly in
//     cmd/pocketmine-go/session.go, outside any Behaviour this package would need).
type Human struct {
	Living

	uuid string
}

// NewHuman is a port of a slice of Human::__construct - see Human's own doc comment for what's
// deliberately left out. uuid is real PHP's UuidInterface, kept as a bare string here (this port
// has no need for a structured UUID type anywhere else yet).
func NewHuman(position math.Vector3, boundingBox math.AxisAlignedBB, uuid string) *Human {
	h := &Human{
		Living: Living{Entity: Entity{position: position, boundingBox: boundingBox, health: 20, maxHealth: 20}},
		uuid:   uuid,
	}
	h.Init(h)
	return h
}

// GetUniqueID is a port of Human::getUniqueId.
func (h *Human) GetUniqueID() string { return h.uuid }

// GetEyeHeight is a port of Entity::getEyeHeight, using Human's own fixed EntitySizeInfo eye
// height (see humanEyeHeight's own doc comment on why this isn't a general Entity mechanism yet).
func (h *Human) GetEyeHeight() float64 { return humanEyeHeight }

// GetEyePos is a port of Entity::getEyePos.
func (h *Human) GetEyePos() math.Vector3 { return h.position.Add(0, h.GetEyeHeight(), 0) }
