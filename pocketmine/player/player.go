package player

import (
	stdmath "math"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/entity"
	"pocketmine-go/pocketmine/inventory"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/world"
)

// var _ block.Player = (*Player)(nil) confirms *Player structurally satisfies block.Player's local
// interface (ResetFallDistance/GetPosition/.../Attack via Entity, GetHorizontalFacing/IsSneaking/
// GetYaw/GetID/GetEyePos/IsSurvival) - the same compile-time check World.go uses for block.World.
var _ block.Player = (*Player)(nil)

// humanBoundingBoxHalfWidth/Height mirror Human::getInitialSizeInfo's EntitySizeInfo(1.8, 0.6, ...)
// - width 0.6 (±0.3 from center), height 1.8.
const (
	humanBoundingBoxHalfWidth = 0.3
	humanBoundingBoxHeight    = 1.8
)

// mainInventorySize mirrors PlayerInventory's own real slot count (9 hotbar + 27 main = 36).
const mainInventorySize = 36

// Player is a port of a slice of pocketmine\player\Player.
//
// This is a deliberately narrow first slice of a 2900+ line class - real Player also covers chunk
// streaming (per-player view distance, ChunkSelector, usedChunks/loadQueue), inventory windows
// beyond the base inventory (cursor/crafting-grid/creative), forms, sleeping, respawn, disk-
// persisted PlayerInfo, and SurvivalBlockBreakHandler's precise break-time state machine - none of
// that is attempted here. What this slice does provide is real, not stubbed: identity (username/
// UUID/XUID), a real main Inventory (see NewHuman's own doc comment on why it lives here instead
// of on entity.Human), GameMode, and everything block.Player/block.Living/block.Entity's local
// interfaces need to treat a *Player as a genuine entity registered in a World - the previous
// stand-in (cmd/pocketmine-go's own "session" struct) can now hold one of these instead of
// reimplementing player-shaped state itself.
type Player struct {
	entity.Human

	// id is this player's entity ID (block.Player.GetID(), also needed to satisfy World's own
	// registeredEntity interface for AddEntity/RemoveEntity) - *entity.Entity itself has no GetID
	// of its own (see explosion_test.go's identical explosionTestEntity wrapper, needed for the
	// same reason), so Player supplies one directly instead. Caller-assigned at construction,
	// matching how cmd/pocketmine-go's own connection-accept path already hands out a unique
	// entity runtime ID per session.
	id int

	username    string
	displayName string
	xuid        string
	gameMode    GameMode
	spawned     bool

	world         *world.World
	spawnPosition math.Vector3

	yaw, pitch float64
	flying     bool

	inventory *inventory.SimpleInventory
}

// NewPlayer is a port of a slice of Player::__construct/PlayerInfo - see Player's own doc comment
// for what's deliberately left out. id is this player's entity ID (see the Player.id field's own
// doc comment on why the caller supplies it rather than Player generating one itself).
func NewPlayer(id int, username, uuid, xuid string, w *world.World, position math.Vector3, gameMode GameMode) *Player {
	bb, _ := math.NewAxisAlignedBB(
		position.X-humanBoundingBoxHalfWidth, position.Y, position.Z-humanBoundingBoxHalfWidth,
		position.X+humanBoundingBoxHalfWidth, position.Y+humanBoundingBoxHeight, position.Z+humanBoundingBoxHalfWidth,
	)

	p := &Player{
		Human:       *entity.NewHuman(position, bb, uuid),
		id:          id,
		username:    username,
		displayName: username,
		xuid:        xuid,
		gameMode:    gameMode,
		world:       w,
		inventory:   inventory.NewSimpleInventory(mainInventorySize),
	}
	return p
}

// GetID is a port of Entity::getId - see the Player.id field's own doc comment on why Player
// defines this itself rather than inheriting it.
func (p *Player) GetID() int { return p.id }

// GetName is a port of Player::getName (IPlayer/OfflinePlayer's shared getName - real PHP has
// several overlapping name getters across Player's interfaces; this port only needs one).
func (p *Player) GetName() string { return p.username }

// GetDisplayName is a port of Player::getDisplayName.
func (p *Player) GetDisplayName() string { return p.displayName }

// SetDisplayName is a port of Player::setDisplayName.
func (p *Player) SetDisplayName(name string) { p.displayName = name }

// GetXuid is a port of Player::getXuid.
func (p *Player) GetXuid() string { return p.xuid }

// GetWorld is a port of Human::getWorld (declared on Living in real PHP's Location, simplified
// here to a bare accessor - this port's Human/Entity doesn't carry a Location object, just a bare
// Vector3 position, so the owning World is tracked directly on Player instead).
func (p *Player) GetWorld() *world.World { return p.world }

// IsSpawned is a port of Player::$spawned (there's no dedicated getter in real PHP - the property
// itself is public - but this port keeps its fields unexported, matching its own convention
// elsewhere).
func (p *Player) IsSpawned() bool { return p.spawned }

// SetSpawned marks this player as spawned - a port of the several `$this->spawned = true;`
// assignments scattered through Player::sendChunk/doFirstSpawn (this port has no chunk-send state
// machine to hook that transition to yet, so callers set this directly once whatever spawn
// sequence this port does end up implementing decides the player is ready).
func (p *Player) SetSpawned(spawned bool) { p.spawned = spawned }

// GetInventory is a port of Human::getInventory - see NewHuman's own doc comment on why this
// lives on Player instead of entity.Human.
func (p *Player) GetInventory() *inventory.SimpleInventory { return p.inventory }

// GetGamemode is a port of Player::getGamemode.
func (p *Player) GetGamemode() GameMode { return p.gameMode }

// SetGamemode is a port of a slice of Player::setGamemode - minus the cancellable
// PlayerGameModeChangeEvent and ability-recalculation/effect-clearing side effects real PHP
// applies when switching modes (no event bus or AbilityMap/EffectManager wired up here yet).
func (p *Player) SetGamemode(gameMode GameMode) { p.gameMode = gameMode }

// IsSurvival is a port of Player::isSurvival($literal = false) - block.Player's local interface
// only needs the non-literal form (Adventure counts as survival-like for block-breaking purposes).
func (p *Player) IsSurvival() bool {
	return p.gameMode == GameModeSurvival || p.gameMode == GameModeAdventure
}

// IsCreative is a port of Player::isCreative($literal = false).
func (p *Player) IsCreative() bool {
	return p.gameMode == GameModeCreative || p.gameMode == GameModeSpectator
}

// IsAdventure is a port of Player::isAdventure($literal = false).
func (p *Player) IsAdventure() bool {
	return p.gameMode == GameModeAdventure || p.gameMode == GameModeSpectator
}

// IsSpectator is a port of Player::isSpectator.
func (p *Player) IsSpectator() bool { return p.gameMode == GameModeSpectator }

// GetSpawn is a port of Player::getSpawn (falls back to the world's own spawn if none is set for
// this player specifically, matching real PHP's `$this->spawnPosition ?? $world->getSpawnLocation()`
// via requestSafeSpawn/getSpawn's null-coalescing default).
func (p *Player) GetSpawn() math.Vector3 {
	if p.spawnPosition == (math.Vector3{}) {
		return p.world.GetSpawnLocation()
	}
	return p.spawnPosition
}

// SetSpawn is a port of Player::setSpawn.
func (p *Player) SetSpawn(pos math.Vector3) { p.spawnPosition = pos }

// GetYaw/GetPitch/SetRotation port Location's yaw/pitch accessors as used by Player.
func (p *Player) GetYaw() float64   { return p.yaw }
func (p *Player) GetPitch() float64 { return p.pitch }
func (p *Player) SetRotation(yaw, pitch float64) {
	p.yaw = yaw
	p.pitch = pitch
}

// GetHorizontalFacing is a port of Entity::getHorizontalFacing.
func (p *Player) GetHorizontalFacing() math.Facing {
	angle := stdmath.Mod(p.yaw, 360)
	if angle < 0 {
		angle += 360
	}

	switch {
	case (angle >= 0 && angle < 45) || (angle >= 315 && angle < 360):
		return math.South
	case angle >= 45 && angle < 135:
		return math.West
	case angle >= 135 && angle < 225:
		return math.North
	default:
		return math.East
	}
}

// IsFlying is a port of Player::isFlying.
func (p *Player) IsFlying() bool { return p.flying }

// SetFlying is a port of a slice of Player::setFlying - minus real PHP's ability-sync-to-client
// side effect (no AbilityMap/network session type exists in this package to sync to).
func (p *Player) SetFlying(flying bool) { p.flying = flying }
