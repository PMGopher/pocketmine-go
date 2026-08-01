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

// ItemUseOnBlockSound is a port of pocketmine\world\sound\ItemUseOnBlockSound.
//
// The PHP original stores the whole Block (its Encode() only ever reads getStateId() from it).
// Storing block.Behavior here instead would create an import cycle (block already imports this
// package for RedstonePowerOnSound/OffSound), so this stores just the state ID that Encode()
// actually needs.
type ItemUseOnBlockSound struct {
	BlockStateID int
}

// AnvilFallSound is a port of pocketmine\world\sound\AnvilFallSound.
type AnvilFallSound struct{}

// CopperWaxApplySound is a port of pocketmine\world\sound\CopperWaxApplySound.
type CopperWaxApplySound struct{}

// CopperWaxRemoveSound is a port of pocketmine\world\sound\CopperWaxRemoveSound.
type CopperWaxRemoveSound struct{}

// ScrapeSound is a port of pocketmine\world\sound\ScrapeSound.
type ScrapeSound struct{}

// DoorSound is a port of pocketmine\world\sound\DoorSound.
type DoorSound struct{}

// PressurePlateActivateSound is a port of pocketmine\world\sound\PressurePlateActivateSound.
// Stores the block's state ID rather than the whole Block - same reasoning as ItemUseOnBlockSound.
type PressurePlateActivateSound struct {
	BlockStateID int
}

// PressurePlateDeactivateSound is a port of pocketmine\world\sound\PressurePlateDeactivateSound.
type PressurePlateDeactivateSound struct {
	BlockStateID int
}

// AmethystBlockChimeSound is a port of pocketmine\world\sound\AmethystBlockChimeSound.
type AmethystBlockChimeSound struct{}

// BlockPunchSound is a port of pocketmine\world\sound\BlockPunchSound. Stores the block's state
// ID rather than the whole Block - same reasoning as ItemUseOnBlockSound.
type BlockPunchSound struct {
	BlockStateID int
}

// BlazeShootSound is a port of pocketmine\world\sound\BlazeShootSound.
type BlazeShootSound struct{}

// FireExtinguishSound is a port of pocketmine\world\sound\FireExtinguishSound.
type FireExtinguishSound struct{}

// FlintSteelSound is a port of pocketmine\world\sound\FlintSteelSound.
type FlintSteelSound struct{}
