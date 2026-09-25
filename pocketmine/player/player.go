package player

import (
	"time"

	"github.com/google/uuid"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/entity"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
	"pocketmine-go/pocketmine/world"
)

// var _ block.Player = (*Player)(nil) confirms *Player structurally satisfies block.Player's local
// interface - the same compile-time check World.go uses for block.World.
var _ block.Player = (*Player)(nil)

// var _ entity.Player = (*Player)(nil) confirms *Player satisfies the entity package's view of a
// player (pickups, collision callbacks, game-mode checks).
var _ entity.Player = (*Player)(nil)

// Player is a port of a large slice of pocketmine\player\Player (2900+ lines in the original),
// built on a real entity.Human: health, hunger, experience, effects, armor/offhand/ender
// inventories, damage, knockback, death and network (de)spawning all come from the entity
// package, exactly as in PHP where Player extends Human.
//
// Also real: identity (username/UUID/XUID/firstPlayed/lastPlayed), GameMode,
// flight/auto-jump/block-collision/sneak-pressed flags, per-player chunk streaming (view distance,
// ChunkSelector-driven load ordering, UsedChunkStatus tracking, world.ChunkListener - see
// chunk_streaming.go), survival block-breaking (see block_interaction.go/
// survival_block_break_handler.go), PvP (AttackEntity - see combat.go), fall damage (see
// fall_damage.go) and the Player-specific entity overrides in entity_overrides.go (ticking
// without server-side physics, picking up items/arrows, damage rules, death).
//
// Not ported (each needs a subsystem this port doesn't have yet): NetworkSession (packets go
// through SetPacketSender), inventory windows beyond the player's own inventories, forms,
// sleeping/respawn, PlayerInfo/PlayerDataProvider persistence wiring, permissions/CommandSender,
// chat broadcasting, item use/consumption actions, server-side movement validation (position is
// trusted from the client's PlayerAuthInput reports), and the pocketmine\event\player events
// other than PlayerExhaustEvent/PlayerExperienceChangeEvent.
type Player struct {
	entity.Human

	username    string
	displayName string
	xuid        string
	playerUUID  uuid.UUID
	gameMode    GameMode
	spawned     bool

	spawnPosition math.Vector3

	flying bool

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
	viewDistance int
	usedChunks   map[[2]int]UsedChunkStatus
	loadQueue    map[[2]int]bool
	// loadQueueOrder is loadQueue's iteration order (PHP arrays keep insertion order): nearest
	// chunk first, as orderChunks fills it.
	loadQueueOrder [][2]int
	// nextChunkOrderRun is Player::$nextChunkOrderRun: ticks until OrderChunks runs again
	// (noChunkOrderRun when nothing needs reordering).
	nextChunkOrderRun int
	// chunksPerTick is Player::$chunksPerTick (pocketmine.yml chunk-sending.per-tick, default 4).
	chunksPerTick int
	tickingChunks map[[2]int]bool

	// packetSender backs SendPacket/SetPacketSender (see network.go) - real PHP reaches this
	// player's NetworkSession directly; this port has no NetworkSession type, so the caller that
	// owns the actual connection (cmd/pocketmine-go) supplies this closure once instead.
	packetSender PacketSender

	// server and networkSession are PHP's $this->server and $this->networkSession, as the small
	// interfaces this package needs from them (see server.go). Both are nil in tests.
	server         Server
	networkSession NetworkSession
	// messageCounter is Player::$messageCounter: chat messages still allowed this tick.
	messageCounter int
	// startAction is Player::$startAction: the server tick item use started on, or -1.
	startAction int64

	// lastBroadcastLocation mirrors the lastLocation Player::processMostRecentMovements compares
	// against to decide whether to broadcast movement.
	lastBroadcastLocation entity.Location

	inAirTicks int
}

// NewPlayer is a port of Player::__construct (for a player with no saved data): the player is
// created as a real entity in w at position, with its own runtime entity ID.
func NewPlayer(username string, playerUUID uuid.UUID, xuid string, w *world.World, position math.Vector3, gameMode GameMode, skin *entity.Skin) *Player {
	return NewPlayerFromData(username, playerUUID, xuid, entity.LocationFromObject(position, w, 0, 0), gameMode, true, skin, nil, nil)
}

// NewPlayerFromData is a port of Player::__construct with saved player data ($namedtag, from
// Server::getOfflinePlayerData) plus the data half of Player::initEntity: first/last played, the
// saved game mode (unless forceGameMode, like server.properties' force-gamemode) and the custom
// spawn point. tag may be nil for a new player. worldByName resolves SpawnLevel (PHP's
// WorldManager::getWorldByName) and may be nil.
func NewPlayerFromData(username string, playerUUID uuid.UUID, xuid string, location entity.Location, gameMode GameMode, forceGameMode bool, skin *entity.Skin, tag *nbt.CompoundTag, worldByName func(name string) (*world.World, bool)) *Player {
	now := time.Now().UnixMilli()
	p := &Player{
		username:              username,
		displayName:           username,
		xuid:                  xuid,
		playerUUID:            playerUUID,
		gameMode:              gameMode,
		firstPlayed:           now,
		lastPlayed:            now,
		hasBlockCollision:     true,
		autoJump:              true,
		flightSpeedMultiplier: DefaultFlightSpeedMultiplier,
		viewDistance:          -1,
		messageCounter:        2,
		startAction:           -1,
		usedChunks:            map[[2]int]UsedChunkStatus{},
		loadQueue:             map[[2]int]bool{},
		chunksPerTick:         4,
		nextChunkOrderRun:     5,
		tickingChunks:         map[[2]int]bool{},
	}
	p.ConstructHuman(p, location, skin, tag)

	if tag != nil {
		p.firstPlayed = int64(tag.GetLongOr(TagFirstPlayed, nbt.LongTag(now)))
		p.lastPlayed = int64(tag.GetLongOr(TagLastPlayed, nbt.LongTag(now)))
		if gameModeTag, ok := tag.GetTag(tagGameMode); ok && !forceGameMode {
			if id, ok := gameModeTag.(nbt.IntTag); ok {
				gameMode = GameModeSurvival //TODO: bad hack here to avoid crashes on corrupted data
				if id >= nbt.IntTag(GameModeSurvival) && id <= nbt.IntTag(GameModeSpectator) {
					gameMode = GameMode(id)
				}
			}
		}
		if worldByName != nil {
			if w, ok := worldByName(string(tag.GetStringOr(tagSpawnWorld, ""))); ok && w != nil {
				p.spawnPosition = math.NewVector3(float64(tag.GetIntOr(tagSpawnX, 0)), float64(tag.GetIntOr(tagSpawnY, 0)), float64(tag.GetIntOr(tagSpawnZ, 0)))
			}
		}
	}
	p.internalSetGameMode(gameMode)
	p.lastBroadcastLocation = p.GetLocation()
	p.initNetworkHooks()
	return p
}

// Saved player data tag names, Player::TAG_* (TagFirstPlayed/TagLastPlayed are in offline_player.go).
const (
	tagGameMode      = "playerGameType"
	tagSpawnWorld    = "SpawnLevel"
	tagSpawnX        = "SpawnX"
	tagSpawnY        = "SpawnY"
	tagSpawnZ        = "SpawnZ"
	TagLevel         = "Level"
	TagLastKnownXUID = "LastKnownXUID"
)

// GetSaveData is a port of Player::getSaveData: what Server::saveOfflinePlayerData writes to
// players/<name>.dat. The death position isn't saved (Player::$deathPosition isn't ported).
func (p *Player) GetSaveData() *nbt.CompoundTag {
	tag := p.SaveNBT()

	tag.SetString(TagLastKnownXUID, nbt.StringTag(p.xuid))

	if w := p.GetWorld(); w != nil {
		tag.SetString(TagLevel, nbt.StringTag(w.GetFolderName()))
	}

	if p.spawnPosition != (math.Vector3{}) {
		tag.SetString(tagSpawnWorld, nbt.StringTag(p.GetWorld().GetFolderName()))
		tag.SetInt(tagSpawnX, nbt.IntTag(p.spawnPosition.FloorX()))
		tag.SetInt(tagSpawnY, nbt.IntTag(p.spawnPosition.FloorY()))
		tag.SetInt(tagSpawnZ, nbt.IntTag(p.spawnPosition.FloorZ()))
	}

	tag.SetInt(tagGameMode, nbt.IntTag(p.gameMode))
	tag.SetLong(TagFirstPlayed, nbt.LongTag(p.firstPlayed))
	tag.SetLong(TagLastPlayed, nbt.LongTag(time.Now().UnixMilli()))

	return tag
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

// GetName is a port of Player::getName (IPlayer/OfflinePlayer's shared getName - real PHP has
// several overlapping name getters across Player's interfaces; this port only needs one).
func (p *Player) GetName() string { return p.username }

// GetYaw is Player's shorthand for getLocation()->getYaw() (block.Player needs it).
func (p *Player) GetYaw() float64 { return p.GetLocation().Yaw }

// GetPitch is Player's shorthand for getLocation()->getPitch().
func (p *Player) GetPitch() float64 { return p.GetLocation().Pitch }

// GetDisplayName is a port of Player::getDisplayName.
func (p *Player) GetDisplayName() string { return p.displayName }

// SetDisplayName is a port of Player::setDisplayName.
func (p *Player) SetDisplayName(name string) { p.displayName = name }

// GetXuid is a port of Player::getXuid.
func (p *Player) GetXuid() string { return p.xuid }

// IsSpawned is a port of Player::$spawned (there's no dedicated getter in real PHP - the property
// itself is public - but this port keeps its fields unexported, matching its own convention
// elsewhere).
func (p *Player) IsSpawned() bool { return p.spawned }

// SetSpawned marks this player as spawned. Becoming spawned is the entity half of
// Player::doFirstSpawn: the player is spawned to everyone who can see it, and every entity in the
// chunks it has received is spawned to it.
//
// SetSpawned marks this player as spawned - a port of the several `$this->spawned = true;`
// assignments scattered through Player::sendChunk/doFirstSpawn (this port has no chunk-send state
// machine to hook that transition to yet, so callers set this directly once whatever spawn
// sequence this port does end up implementing decides the player is ready).
func (p *Player) SetSpawned(spawned bool) {
	wasSpawned := p.spawned
	p.spawned = spawned
	if spawned && !wasSpawned {
		p.SpawnToAll()
		p.spawnEntitiesOnAllChunks()
	}
}

// GetGamemode is a port of Player::getGamemode.
func (p *Player) GetGamemode() GameMode { return p.gameMode }

// internalSetGameMode is a port of Player::internalSetGameMode. Not ported: the spectator
// sendPosition(MODE_TELEPORT) resync and the checkGroundState(0,...) call (both need the network
// session/server-side movement handling this port doesn't have).
func (p *Player) internalSetGameMode(gameMode GameMode) {
	p.gameMode = gameMode

	p.allowFlight = p.gameMode == GameModeCreative
	p.GetHungerManager().SetEnabled(p.IsSurvival())

	if p.IsSpectator() {
		p.SetFlying(true)
		p.SetHasBlockCollision(false)
		p.SetSilent(true)
		p.OnGround = false
	} else {
		if p.IsSurvival() {
			p.SetFlying(false)
		}
		p.SetHasBlockCollision(true)
		p.SetSilent(false)
	}
}

// SetGamemode is a port of Player::setGamemode, minus the cancellable PlayerGameModeChangeEvent
// (not ported) and syncing the game mode to the client (no network session - the caller must send
// it). Returns whether the game mode changed.
func (p *Player) SetGamemode(gameMode GameMode) bool {
	if p.gameMode == gameMode {
		return false
	}

	p.internalSetGameMode(gameMode)

	if p.IsSpectator() {
		p.DespawnFromAll()
	} else {
		p.SpawnToAll()
	}
	return true
}

// HasFiniteResources is a port of Player::hasFiniteResources: whether the player's game mode
// consumes items (everything but creative).
func (p *Player) HasFiniteResources() bool { return p.gameMode != GameModeCreative }

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
		return p.GetWorld().GetSpawnLocation()
	}
	return p.spawnPosition
}

// SetSpawn is a port of Player::setSpawn.
func (p *Player) SetSpawn(pos math.Vector3) { p.spawnPosition = pos }

// IsFlying is a port of Player::isFlying.
func (p *Player) IsFlying() bool { return p.flying }

// SetFlying is a port of a slice of Player::setFlying - minus real PHP's ability-sync-to-client
// side effect (no AbilityMap/network session type exists in this package to sync to).
func (p *Player) SetFlying(flying bool) { p.flying = flying }
