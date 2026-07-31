// Package sound is a port of pocketmine\world\sound.
package sound

// Sound is a port of pocketmine\world\sound\Sound. The real interface has one method,
// Encode(pos Vector3) []ClientboundPacket, but ClientboundPacket belongs to the unported
// network/mcpe/protocol package — deferred the same way block.Tile is, as a marker interface
// until that package exists.
type Sound interface {
	//marker — Encode(pos math.Vector3) []ClientboundPacket once network/mcpe/protocol is ported
}

// RedstonePowerOnSound is a port of pocketmine\world\sound\RedstonePowerOnSound.
type RedstonePowerOnSound struct{}

// RedstonePowerOffSound is a port of pocketmine\world\sound\RedstonePowerOffSound.
type RedstonePowerOffSound struct{}
