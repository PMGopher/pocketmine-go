package player

import (
	stdmath "math"
	"time"

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

// Player is a port of a large slice of pocketmine\player\Player (2900+ lines in the original).
//
// Real, not stubbed: identity (username/UUID/XUID/firstPlayed/lastPlayed), a real main Inventory
// (see NewHuman's own doc comment on why it lives here instead of on entity.Human), GameMode,
// flight/auto-jump/block-collision/sneak-pressed flags, real per-player chunk streaming (view
// distance, ChunkSelector-driven load ordering, UsedChunkStatus tracking, world.ChunkListener -
// see chunk_streaming.go), and real survival block-breaking (AttackBlock/ContinueBreakBlock/
// StopBreakBlock/BreakBlock/UpdateBreakingBlock driving a genuine SurvivalBlockBreakHandler - see
// block_interaction.go) - everything block.Player/block.Living/block.Entity's local interfaces
// need to treat a *Player as a genuine entity registered in a World. cmd/pocketmine-go's own
// "session" struct (previously an explicitly-documented stand-in for exactly this) now wraps one
// of these instead of reimplementing player-shaped state itself.
//
// Not ported (each needs a real subsystem this port doesn't have anywhere else yet either, so
// each is a documented gap, not a guess - see the individual methods' own doc comments for exact
// PHP-behaviour differences): inventory windows beyond the base inventory (cursor/crafting-grid/
// creative), forms, sleeping/respawn, PlayerInfo/PlayerDataProvider aren't wired into
// NewPlayer/persistence automatically yet (both real types exist and are usable, just not
// connected to a save/load pipeline), permissions/CommandSender, chat (ChatFormatter exists and is
// usable, just not wired to a broadcast pipeline), item use/consumption, entity attack/interact,
// hunger/experience, and the cancellable events real PHP fires throughout (no event bus wired to
// World/Player - matches every other "no event bus yet" gap elsewhere in this port).
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

	// blockBreakHandler mirrors Player::$blockBreakHandler - non-nil exactly while a survival
	// block-break action is in progress (see block_interaction.go's AttackBlock/StopBreakBlock/
	// UpdateBreakingBlock).
	blockBreakHandler *SurvivalBlockBreakHandler

	// firstPlayed/lastPlayed mirror Player::$firstPlayed/$lastPlayed (IPlayer's own
	// getFirstPlayed/getLastPlayed) - Unix milliseconds, matching real PHP's own
	// `(int) (microtime(true) * 1000)` unit.
	firstPlayed, lastPlayed int64

	// allowFlight/hasBlockCollision/autoJump/flightSpeedMultiplier/sneakPressed mirror the
	// identically-named Player fields.
	allowFlight           bool
	hasBlockCollision     bool
	autoJump              bool
	flightSpeedMultiplier float64
	sneakPressed          bool

	locale string

	// viewDistance/usedChunks/loadQueue/tickingChunks back OrderChunks/RequestChunks and the
	// world.ChunkListener implementation in chunk_streaming.go - see OrderChunks' own doc comment
	// on the -1 default.
	viewDistance  int
	usedChunks    map[[2]int]UsedChunkStatus
	loadQueue     map[[2]int]bool
	tickingChunks map[[2]int]bool
}

// NewPlayer is a port of a slice of Player::__construct/PlayerInfo - see Player's own doc comment
// for what's deliberately left out. id is this player's entity ID (see the Player.id field's own
// doc comment on why the caller supplies it rather than Player generating one itself).
func NewPlayer(id int, username, uuid, xuid string, w *world.World, position math.Vector3, gameMode GameMode) *Player {
	bb, _ := math.NewAxisAlignedBB(
		position.X-humanBoundingBoxHalfWidth, position.Y, position.Z-humanBoundingBoxHalfWidth,
		position.X+humanBoundingBoxHalfWidth, position.Y+humanBoundingBoxHeight, position.Z+humanBoundingBoxHalfWidth,
	)

	now := time.Now().UnixMilli()
	p := &Player{
		Human:                 *entity.NewHuman(position, bb, uuid),
		id:                    id,
		username:              username,
		displayName:           username,
		xuid:                  xuid,
		gameMode:              gameMode,
		world:                 w,
		inventory:             inventory.NewSimpleInventory(mainInventorySize),
		firstPlayed:           now,
		lastPlayed:            now,
		hasBlockCollision:     true,
		autoJump:              true,
		flightSpeedMultiplier: DefaultFlightSpeedMultiplier,
		viewDistance:          -1,
		usedChunks:            map[[2]int]UsedChunkStatus{},
		loadQueue:             map[[2]int]bool{},
		tickingChunks:         map[[2]int]bool{},
	}
	return p
}

// DefaultFlightSpeedMultiplier mirrors Player::DEFAULT_FLIGHT_SPEED_MULTIPLIER.
const DefaultFlightSpeedMultiplier = 0.05

// GetFirstPlayed/GetLastPlayed/HasPlayedBefore port IPlayer's own methods (see the Player.
// firstPlayed/lastPlayed fields' doc comment) - a connected Player always "has played before" by
// the time it exists, matching real PHP's own constructor always setting both to a real value
// (loaded from NBT, or "now" if this is a first join).
func (p *Player) GetFirstPlayed() (int64, bool) { return p.firstPlayed, true }
func (p *Player) GetLastPlayed() (int64, bool)  { return p.lastPlayed, true }
func (p *Player) HasPlayedBefore() bool         { return true }

// SetFirstPlayed/SetLastPlayed let a caller restore these from previously-saved player data (see
// PlayerDataProvider) instead of the "now" default NewPlayer otherwise applies.
func (p *Player) SetFirstPlayed(firstPlayed int64) { p.firstPlayed = firstPlayed }
func (p *Player) SetLastPlayed(lastPlayed int64)   { p.lastPlayed = lastPlayed }

// GetAllowFlight/SetAllowFlight port Player::getAllowFlight/setAllowFlight - minus the ability-sync
// side effect (no AbilityMap/network session type exists in this package to sync to).
func (p *Player) GetAllowFlight() bool      { return p.allowFlight }
func (p *Player) SetAllowFlight(allow bool) { p.allowFlight = allow }

// HasBlockCollision/SetHasBlockCollision port Player::hasBlockCollision/setHasBlockCollision.
func (p *Player) HasBlockCollision() bool         { return p.hasBlockCollision }
func (p *Player) SetHasBlockCollision(value bool) { p.hasBlockCollision = value }

// HasAutoJump/SetAutoJump port Player::hasAutoJump/setAutoJump - minus the ability-sync side
// effect (see GetAllowFlight's own doc comment for the same reason).
func (p *Player) HasAutoJump() bool      { return p.autoJump }
func (p *Player) SetAutoJump(value bool) { p.autoJump = value }

// GetFlightSpeedMultiplier/SetFlightSpeedMultiplier port Player::getFlightSpeedMultiplier/
// setFlightSpeedMultiplier - minus the ability-sync side effect and the real
// InvalidArgumentException on a non-finite/negative value (this port trusts callers, matching its
// own "no shortcuts, but no unreachable-defensive-checks either" convention).
func (p *Player) GetFlightSpeedMultiplier() float64           { return p.flightSpeedMultiplier }
func (p *Player) SetFlightSpeedMultiplier(multiplier float64) { p.flightSpeedMultiplier = multiplier }

// IsSneakPressed/SetSneakPressed port Player::isSneakPressed/setSneakPressed.
func (p *Player) IsSneakPressed() bool         { return p.sneakPressed }
func (p *Player) SetSneakPressed(pressed bool) { p.sneakPressed = pressed }

// GetLocale/SetLocale port PlayerInfo::getLocale (Player itself just delegates to its PlayerInfo
// in real PHP - this port stores it directly on Player instead, since PlayerInfo here is a
// separate, optional value type rather than something Player always carries one of).
func (p *Player) GetLocale() string       { return p.locale }
func (p *Player) SetLocale(locale string) { p.locale = locale }

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
