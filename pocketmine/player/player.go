package player

import (
	"fmt"
	"pocketmine-go/pocketmine/item"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/command"
	"pocketmine-go/pocketmine/entity"
	"pocketmine-go/pocketmine/event"
	playerevent "pocketmine-go/pocketmine/event/player"
	"pocketmine-go/pocketmine/form"
	"pocketmine-go/pocketmine/inventory"
	"pocketmine-go/pocketmine/lang"
	"pocketmine-go/pocketmine/log"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/nbt"
	"pocketmine-go/pocketmine/permission"
	"pocketmine-go/pocketmine/utils"
	"pocketmine-go/pocketmine/world"
)

var _ block.Player = (*Player)(nil)
var _ entity.Player = (*Player)(nil)
var _ IPlayer = (*Player)(nil)

// Player constants, Player::*.
const (
	movesPerTick    = 2
	moveBacklogSize = 100 * movesPerTick //100 ticks backlog (5 seconds)

	// MaxChatCharLength is the max length of a chat message (UTF-8 codepoints, not bytes).
	MaxChatCharLength = 512
	// MaxChatByteLength is the max length of a chat message in bytes. This is a theoretical
	// maximum (if every character was 4 bytes).
	MaxChatByteLength = MaxChatCharLength * 4

	maxReachDistanceCreative          = 13
	maxReachDistanceSurvival          = 7
	maxReachDistanceEntityInteraction = 8

	// DefaultFlightSpeedMultiplier is Player::DEFAULT_FLIGHT_SPEED_MULTIPLIER.
	DefaultFlightSpeedMultiplier = 0.05
)

// Saved player data tag names, Player::TAG_* (TagFirstPlayed/TagLastPlayed are in offline_player.go).
const (
	tagGameMode      = "playerGameType"
	tagSpawnWorld    = "SpawnLevel"
	tagSpawnX        = "SpawnX"
	tagSpawnY        = "SpawnY"
	tagSpawnZ        = "SpawnZ"
	tagDeathWorld    = "DeathLevel"
	tagDeathX        = "DeathPositionX"
	tagDeathY        = "DeathPositionY"
	tagDeathZ        = "DeathPositionZ"
	TagLevel         = "Level"
	TagLastKnownXUID = "LastKnownXUID"
)

var invalidUserNameChars = regexp.MustCompile(`[^A-Za-z0-9_ ]`)

// IsValidUserName is a port of Player::isValidUserName.
func IsValidUserName(name string) bool {
	lname := strings.ToLower(name)
	return lname != "rcon" && lname != "console" && len(name) >= 1 && len(name) <= 16 && !invalidUserNameChars.MatchString(name)
}

// Position is pocketmine\world\Position for the spawn and death points: a position in a world.
type Position struct {
	math.Vector3
	World *world.World
}

// IsValid is Position::isValid: the world is still loaded.
func (p *Position) IsValid() bool {
	return p != nil && p.World != nil && !p.World.IsClosed()
}

// Player is a port of pocketmine\player\Player: the entity of a connected player. It embeds
// entity.Human (PHP: Player extends Human), and is a command sender (CommandSender), a chunk
// listener and an IPlayer.
type Player struct {
	entity.Human
	*permission.Permissible

	server         Server
	networkSession NetworkSession
	logger         log.Logger

	spawned bool

	username      string
	displayName   string
	xuid          string
	authenticated bool
	playerInfo    Info

	currentWindow     inventory.Inventory
	permanentWindows  []inventory.Inventory
	cursorInventory   *inventory.PlayerCursorInventory
	craftingGrid      inventory.TemporaryInventory
	creativeInventory *inventory.CreativeInventory

	messageCounter int

	firstPlayed int64
	lastPlayed  int64
	gamemode    GameMode

	usedChunks map[[2]int]UsedChunkStatus
	// activeChunkGenerationRequests is always empty: generation is synchronous in this port
	// (see World.ensurePopulated), so no request is ever outstanding.
	activeChunkGenerationRequests map[[2]int]bool
	loadQueue                     map[[2]int]bool
	// loadQueueOrder is loadQueue's iteration order (PHP arrays keep insertion order): nearest
	// chunk first, as orderChunks fills it.
	loadQueueOrder    [][2]int
	nextChunkOrderRun int
	tickingChunks     map[[2]int]bool

	viewDistance        int
	spawnThreshold      int
	spawnChunkLoadCount int
	chunksPerTick       int

	hiddenPlayers map[uuid.UUID]bool

	moveRateLimit       float64
	lastMovementProcess *time.Time

	inAirTicks int

	sleeping      *math.Vector3
	spawnPosition *Position
	respawnLocked bool
	deathPosition *Position

	autoJump              bool
	allowFlight           bool
	blockCollision        bool
	flying                bool
	sneakPressed          bool
	flightSpeedMultiplier float64

	lineHeight *int
	locale     string

	startAction int64

	usedItemsCooldown map[string]int64

	lastEmoteTick int64

	formIDCounter int
	forms         map[int]form.Form

	blockBreakHandler *SurvivalBlockBreakHandler

	// lastBroadcastLocation is Entity::$lastLocation as Player uses it: the location movement was
	// last broadcast from (Player::updateMovement is a no-op, so only this file updates it).
	lastBroadcastLocation entity.Location
}

// NewPlayer is a port of Player::__construct followed by Player::initEntity: creates the player
// for a session in spawnLocation's world, from its saved data (namedtag, nil for a new player).
func NewPlayer(server Server, session NetworkSession, playerInfo Info, authenticated bool, spawnLocation entity.Location, namedtag *nbt.CompoundTag) *Player {
	username := utils.Clean(playerInfo.GetUsername(), true)
	p := &Player{
		server:         server,
		networkSession: session,
		playerInfo:     playerInfo,
		authenticated:  authenticated,
		logger:         log.NewPrefixedLogger(server.GetLogger(), "Player: "+username),

		username:    username,
		displayName: username,
		locale:      playerInfo.GetLocale(),

		messageCounter: 2,

		usedChunks:                    map[[2]int]UsedChunkStatus{},
		activeChunkGenerationRequests: map[[2]int]bool{},
		loadQueue:                     map[[2]int]bool{},
		nextChunkOrderRun:             5,
		tickingChunks:                 map[[2]int]bool{},
		viewDistance:                  -1,

		hiddenPlayers: map[uuid.UUID]bool{},
		moveRateLimit: 10 * movesPerTick,

		autoJump:              true,
		blockCollision:        true,
		flightSpeedMultiplier: DefaultFlightSpeedMultiplier,

		startAction:       -1,
		usedItemsCooldown: map[string]int64{},
		forms:             map[int]form.Form{},
	}
	if xbl, ok := playerInfo.(*XboxLivePlayerInfo); ok {
		p.xuid = xbl.GetXuid()
	}

	p.creativeInventory = inventory.GetCreativeInventory()

	rootPermissions := map[string]bool{permission.RootUser: true}
	if server.IsOp(p.username) {
		rootPermissions[permission.RootOperator] = true
	}
	p.Permissible = permission.NewPermissible(rootPermissions)
	p.chunksPerTick = server.GetPropertyInt("chunk-sending.per-tick", 4)
	spawnRadius := server.GetPropertyInt("chunk-sending.spawn-radius", 4)
	p.spawnThreshold = int(float64(spawnRadius*spawnRadius) * 3.141592653589793)

	w := spawnLocation.GetWorld()
	//load the spawn chunk so we can see the terrain
	xSpawnChunk, zSpawnChunk := spawnLocation.FloorX()>>4, spawnLocation.FloorZ()>>4
	w.RegisterChunkLoader(p, xSpawnChunk, zSpawnChunk)
	w.RegisterChunkListener(p, xSpawnChunk, zSpawnChunk)
	p.usedChunks[[2]int{xSpawnChunk, zSpawnChunk}] = UsedChunkStatusNeeded

	p.ConstructHuman(p, spawnLocation, playerInfo.GetSkin(), namedtag)
	p.lastBroadcastLocation = p.GetLocation()
	return p
}

// InitHumanData is a port of Player::initHumanData: the name tag and UUID come from the login.
func (p *Player) InitHumanData(tag *nbt.CompoundTag) {
	p.SetNameTag(p.username)
	if id, err := uuid.Parse(p.playerInfo.GetUUID()); err == nil {
		p.SetUniqueID(id)
	}
}

// callDummyItemHeldEvent is a port of Player::callDummyItemHeldEvent.
func (p *Player) callDummyItemHeldEvent() {
	slot := p.GetInventory().GetHeldItemIndex()
	ev := playerevent.NewPlayerItemHeldEvent(p, p.GetInventory().GetItem(slot), slot)
	event.Call(ev)
	//TODO: this event is actually cancellable, but cancelling it here has no meaningful result, so we
	//just ignore it. We fire this only because the content of the held slot changed, not because the
	//held slot index changed. We can't prevent that from here, and nor would it be sensible to.
}

// InitEntity is a port of Player::initEntity.
func (p *Player) InitEntity(tag *nbt.CompoundTag) {
	p.Human.InitEntity(tag)
	p.addDefaultWindows()

	p.GetInventory().GetListeners().Add(inventory.NewCallbackInventoryListener(
		func(_ inventory.Inventory, slot int, _ item.Item) {
			if slot == p.GetInventory().GetHeldItemIndex() {
				p.SetUsingItem(false)
				p.callDummyItemHeldEvent()
			}
		},
		func(_ inventory.Inventory, _ map[int]item.Item) {
			p.SetUsingItem(false)
			p.callDummyItemHeldEvent()
		},
	))

	now := time.Now().UnixMilli()
	p.firstPlayed = int64(tag.GetLongOr(TagFirstPlayed, nbt.LongTag(now)))
	p.lastPlayed = int64(tag.GetLongOr(TagLastPlayed, nbt.LongTag(now)))

	gameMode := p.server.GetGamemode()
	if gameModeTag, ok := tag.GetTag(tagGameMode); ok && !p.server.GetForceGamemode() {
		if id, ok := gameModeTag.(nbt.IntTag); ok {
			gameMode = GameModeSurvival //TODO: bad hack here to avoid crashes on corrupted data
			if id >= nbt.IntTag(GameModeSurvival) && id <= nbt.IntTag(GameModeSpectator) {
				gameMode = GameMode(id)
			}
		}
	}
	p.internalSetGameMode(gameMode)

	p.KeepMovement = true
	p.SetStepHeight(0.6)

	p.SetNameTagVisible(true)
	p.SetNameTagAlwaysVisible(true)
	p.SetCanClimb(true)

	if w, ok := p.server.GetWorldManager().GetWorldByName(string(tag.GetStringOr(tagSpawnWorld, ""))); ok {
		p.spawnPosition = &Position{Vector3: math.NewVector3(float64(tag.GetIntOr(tagSpawnX, 0)), float64(tag.GetIntOr(tagSpawnY, 0)), float64(tag.GetIntOr(tagSpawnZ, 0))), World: w}
	}
	if w, ok := p.server.GetWorldManager().GetWorldByName(string(tag.GetStringOr(tagDeathWorld, ""))); ok {
		p.deathPosition = &Position{Vector3: math.NewVector3(float64(tag.GetIntOr(tagDeathX, 0)), float64(tag.GetIntOr(tagDeathY, 0)), float64(tag.GetIntOr(tagDeathZ, 0))), World: w}
	}
}

// GetLeaveMessage is a port of Player::getLeaveMessage.
func (p *Player) GetLeaveMessage() any {
	if p.spawned {
		return lang.KnownTranslationFactory.MultiplayerPlayerLeft(p.GetDisplayName()).Prefix(utils.Yellow)
	}
	return ""
}

func (p *Player) IsAuthenticated() bool { return p.authenticated }

// GetPlayerInfo returns an object containing information about the player, such as their
// username, skin, and misc extra client-specific data.
func (p *Player) GetPlayerInfo() Info { return p.playerInfo }

// GetXuid returns the player's Xbox user ID (XUID) if logged into Xbox Live, or "".
func (p *Player) GetXuid() string { return p.xuid }

// GetFirstPlayed is a port of Player::getFirstPlayed.
func (p *Player) GetFirstPlayed() (int64, bool) { return p.firstPlayed, true }

// GetLastPlayed is a port of Player::getLastPlayed.
func (p *Player) GetLastPlayed() (int64, bool) { return p.lastPlayed, true }

// HasPlayedBefore is a port of Player::hasPlayedBefore.
func (p *Player) HasPlayedBefore() bool {
	return p.lastPlayed-p.firstPlayed > 1 // microtime(true) - microtime(true) may have less than one millisecond difference
}

// SetAllowFlight sets whether the player is allowed to toggle flight mode.
func (p *Player) SetAllowFlight(value bool) {
	if p.allowFlight != value {
		p.allowFlight = value
		p.GetNetworkSession().SyncAbilities(p)
	}
}

func (p *Player) GetAllowFlight() bool { return p.allowFlight }

// SetHasBlockCollision sets whether the player's movement may be obstructed by blocks with
// collision boxes.
func (p *Player) SetHasBlockCollision(value bool) {
	if p.blockCollision != value {
		p.blockCollision = value
		p.GetNetworkSession().SyncAbilities(p)
	}
}

func (p *Player) HasBlockCollision() bool { return p.blockCollision }

// SetFlying is a port of Player::setFlying.
func (p *Player) SetFlying(value bool) {
	if p.flying != value {
		p.flying = value
		p.ResetFallDistance()
		p.GetNetworkSession().SyncAbilities(p)
	}
}

func (p *Player) IsFlying() bool { return p.flying }

// SetFlightSpeedMultiplier sets the player's flight speed multiplier. Normal flying speed in
// blocks-per-tick is (multiplier * 10) blocks per tick.
func (p *Player) SetFlightSpeedMultiplier(flightSpeedMultiplier float64) {
	if p.flightSpeedMultiplier != flightSpeedMultiplier {
		p.flightSpeedMultiplier = flightSpeedMultiplier
		p.GetNetworkSession().SyncAbilities(p)
	}
}

func (p *Player) GetFlightSpeedMultiplier() float64 { return p.flightSpeedMultiplier }

func (p *Player) SetAutoJump(value bool) {
	if p.autoJump != value {
		p.autoJump = value
		p.GetNetworkSession().SyncAdventureSettings()
	}
}

func (p *Player) HasAutoJump() bool { return p.autoJump }

// GetServer is a port of Player::getServer (as the command sender's server).
func (p *Player) GetServer() command.Server { return p.server }

// GetPlayerServer returns the server as the player package's Server interface.
func (p *Player) GetPlayerServer() Server { return p.server }

func (p *Player) GetScreenLineHeight() int {
	if p.lineHeight != nil {
		return *p.lineHeight
	}
	return 7
}

// SetScreenLineHeight is a port of Player::setScreenLineHeight (panicking on a height below 1, like
// PHP's InvalidArgumentException).
func (p *Player) SetScreenLineHeight(height *int) {
	if height != nil && *height < 1 {
		panic("Line height must be at least 1")
	}
	p.lineHeight = height
}

// CanSee is a port of Player::canSee.
func (p *Player) CanSee(player *Player) bool { return !p.hiddenPlayers[player.GetUniqueID()] }

// HidePlayer is a port of Player::hidePlayer.
func (p *Player) HidePlayer(player *Player) {
	if player == p {
		return
	}
	p.hiddenPlayers[player.GetUniqueID()] = true
	player.DespawnFrom(p, true)
}

// ShowPlayer is a port of Player::showPlayer.
func (p *Player) ShowPlayer(player *Player) {
	if player == p {
		return
	}
	delete(p.hiddenPlayers, player.GetUniqueID())
	if player.IsOnline() {
		player.SpawnTo(p)
	}
}

// ResetFallDistance is a port of Player::resetFallDistance.
func (p *Player) ResetFallDistance() {
	p.Human.ResetFallDistance()
	p.inAirTicks = 0
}

func (p *Player) GetViewDistance() int { return p.viewDistance }

// SetViewDistance is a port of Player::setViewDistance.
func (p *Player) SetViewDistance(distance int) {
	newViewDistance := p.server.GetAllowedViewDistance(distance)

	if newViewDistance != p.viewDistance {
		event.Call(playerevent.NewPlayerViewDistanceChangeEvent(p, p.viewDistance, newViewDistance))
	}

	p.viewDistance = newViewDistance

	spawnRadius := min(p.viewDistance, p.server.GetPropertyInt("chunk-sending.spawn-radius", 4))
	p.spawnThreshold = int(float64(spawnRadius*spawnRadius) * 3.141592653589793)

	p.nextChunkOrderRun = 0

	p.GetNetworkSession().SyncViewAreaRadius(p.viewDistance)

	p.logger.Debug(fmt.Sprintf("Setting view distance to %d (requested %d)", p.viewDistance, distance))
}

// IsOnline is a port of Player::isOnline.
func (p *Player) IsOnline() bool { return p.IsConnected() }

// IsConnected is a port of Player::isConnected.
func (p *Player) IsConnected() bool {
	return p.networkSession != nil && p.networkSession.IsConnected()
}

// GetNetworkSession is a port of Player::getNetworkSession (panicking like PHP's LogicException
// when the player is not connected).
func (p *Player) GetNetworkSession() NetworkSession {
	if p.networkSession == nil {
		panic("Player is not connected")
	}
	return p.networkSession
}

// GetInvManager is $player->getNetworkSession()->getInvManager(), which inventories use to sync
// their viewers (inventory.NetworkViewer).
func (p *Player) GetInvManager() inventory.InventorySyncer {
	if p.networkSession == nil {
		return nil
	}
	if m := p.networkSession.GetInvManager(); m != nil {
		return m
	}
	return nil
}

// GetName returns the username.
func (p *Player) GetName() string { return p.username }

// GetDisplayName returns the "friendly" display name of this player to use in the chat.
func (p *Player) GetDisplayName() string { return p.displayName }

// SetDisplayName is a port of Player::setDisplayName.
func (p *Player) SetDisplayName(name string) {
	ev := playerevent.NewPlayerDisplayNameChangeEvent(p, p.displayName, name)
	event.Call(ev)
	p.displayName = ev.GetNewName()
}

// CanBeRenamed is a port of Player::canBeRenamed.
func (p *Player) CanBeRenamed() bool { return false }

// GetLocale returns the player's locale, e.g. en_US.
func (p *Player) GetLocale() string { return p.locale }

// GetLanguage is a port of Player::getLanguage.
func (p *Player) GetLanguage() *lang.Language { return p.server.GetLanguage() }

// ChangeSkin is a port of Player::changeSkin: called when a player changes their skin.
func (p *Player) ChangeSkin(skin *entity.Skin, newSkinName, oldSkinName string) bool {
	ev := playerevent.NewPlayerChangeSkinEvent(p, p.GetSkin(), skin)
	event.Call(ev)

	if ev.IsCancelled() {
		p.SendSkin([]world.EntityViewer{p})
		return true
	}

	if newSkin, ok := ev.GetNewSkin().(*entity.Skin); ok {
		p.SetSkin(newSkin)
	}
	p.SendSkin(nil)
	return true
}

// SendSkin is a port of Player::sendSkin: nil targets means every online player (including the
// player itself).
func (p *Player) SendSkin(targets []world.EntityViewer) {
	if targets == nil {
		for _, other := range p.server.GetOnlinePlayers() {
			targets = append(targets, other)
		}
	}
	p.Human.SendSkin(targets)
}

// IsUsingItem returns whether the player is currently using an item (right-click and hold).
func (p *Player) IsUsingItem() bool { return p.startAction > -1 }

// SetUsingItem is a port of Player::setUsingItem.
func (p *Player) SetUsingItem(value bool) {
	p.startAction = -1
	if value {
		p.startAction = p.server.GetTick()
	}
	p.MarkNetworkPropertiesDirty()
}

// GetItemUseDuration returns how long the player has been using their currently-held item for.
// Used for determining arrow shoot force for bows.
func (p *Player) GetItemUseDuration() int64 {
	if p.startAction == -1 {
		return -1
	}
	return p.server.GetTick() - p.startAction
}

// SetPositionInWorld is a port of Player::setPosition: changing worlds unloads the old world's
// chunks and tells the client about the new world.
func (p *Player) SetPositionInWorld(pos math.Vector3, w *world.World) bool {
	var oldWorld *world.World
	if loc := p.GetLocation(); loc.IsValid() {
		oldWorld = loc.GetWorld()
	}
	if p.Human.SetPositionInWorld(pos, w) {
		newWorld := p.GetWorld()
		if oldWorld != newWorld {
			if oldWorld != nil {
				for index := range p.usedChunks {
					p.unloadChunk(index[0], index[1], oldWorld)
				}
			}

			p.usedChunks = map[[2]int]UsedChunkStatus{}
			p.loadQueue = map[[2]int]bool{}
			p.loadQueueOrder = nil
			if p.networkSession != nil {
				p.GetNetworkSession().OnEnterWorld()
			}
		}
		return true
	}
	return false
}

// GetDeathPosition is a port of Player::getDeathPosition.
func (p *Player) GetDeathPosition() *Position {
	if p.deathPosition != nil && !p.deathPosition.IsValid() {
		p.deathPosition = nil
	}
	return p.deathPosition
}

// SetDeathPosition is a port of Player::setDeathPosition: w nil means the player's current world;
// pos nil clears it.
func (p *Player) SetDeathPosition(pos *math.Vector3, w *world.World) {
	if pos != nil {
		if w == nil {
			w = p.GetWorld()
		}
		p.deathPosition = &Position{Vector3: *pos, World: w}
	} else {
		p.deathPosition = nil
	}
	p.MarkNetworkPropertiesDirty()
}

// GetSpawn is a port of Player::getSpawn: the custom spawn point, or the default world's spawn.
func (p *Player) GetSpawn() Position {
	if p.HasValidCustomSpawn() {
		return *p.spawnPosition
	}
	w := p.server.GetWorldManager().GetDefaultWorld()
	return Position{Vector3: w.GetSpawnLocation(), World: w}
}

// HasValidCustomSpawn is a port of Player::hasValidCustomSpawn.
func (p *Player) HasValidCustomSpawn() bool { return p.spawnPosition.IsValid() }

// SetSpawn is a port of Player::setSpawn: sets the spawnpoint of the player (and the compass
// direction); w nil means the player's current world, pos nil clears it.
func (p *Player) SetSpawn(pos *math.Vector3, w *world.World) {
	if pos != nil {
		if w == nil {
			w = p.GetWorld()
		}
		p.spawnPosition = &Position{Vector3: *pos, World: w}
	} else {
		p.spawnPosition = nil
	}
	p.GetNetworkSession().SyncPlayerSpawnPoint(p.GetSpawn().Vector3)
}

// IsSleeping is a port of Player::isSleeping.
func (p *Player) IsSleeping() bool { return p.sleeping != nil }

// bedBlock is block.Bed as Player's sleep methods need it.
type bedBlock interface {
	block.Behavior
	SetOccupied(occupied bool)
}

// SleepOn is a port of Player::sleepOn.
func (p *Player) SleepOn(pos math.Vector3) bool {
	pos = pos.Floor()
	b := p.GetWorld().GetBlock(pos)

	ev := playerevent.NewPlayerBedEnterEvent(p, b)
	event.Call(ev)
	if ev.IsCancelled() {
		return false
	}

	if bed, ok := b.(bedBlock); ok {
		bed.SetOccupied(true)
		_ = p.GetWorld().SetBlock(block.NewPosition(pos.X, pos.Y, pos.Z, p.GetWorld()), bed)
	}

	p.sleeping = &pos
	p.MarkNetworkPropertiesDirty()

	p.SetSpawn(&pos, nil)

	p.GetWorld().SetSleepTicks(60)

	return true
}

// StopSleep is a port of Player::stopSleep.
func (p *Player) StopSleep() {
	if p.sleeping == nil {
		return
	}
	pos := *p.sleeping
	b := p.GetWorld().GetBlock(pos)
	if bed, ok := b.(bedBlock); ok {
		bed.SetOccupied(false)
		_ = p.GetWorld().SetBlock(block.NewPosition(pos.X, pos.Y, pos.Z, p.GetWorld()), bed)
	}
	event.Call(playerevent.NewPlayerBedLeaveEvent(p, b))

	p.sleeping = nil
	p.MarkNetworkPropertiesDirty()

	p.GetWorld().SetSleepTicks(0)

	p.GetNetworkSession().SendDataPacket(&packet.Animate{ActionType: packet.AnimateActionStopSleep, EntityRuntimeID: uint64(p.GetID())})
}

func (p *Player) GetGamemode() GameMode { return p.gamemode }

// internalSetGameMode is a port of Player::internalSetGameMode.
func (p *Player) internalSetGameMode(gameMode GameMode) {
	p.gamemode = gameMode

	p.allowFlight = p.gamemode == GameModeCreative
	p.GetHungerManager().SetEnabled(p.IsSurvival())

	if p.IsSpectator() {
		p.SetFlying(true)
		p.SetHasBlockCollision(false)
		p.SetSilent(true)
		p.OnGround = false

		//TODO: HACK! this syncs the onground flag with the client so that flying works properly
		//this is a yucky hack but we don't have any other options :(
		loc := p.GetLocation()
		p.sendPosition(loc.Vector3, nil, nil, packet.MoveModeTeleport)
	} else {
		if p.IsSurvival() {
			p.SetFlying(false)
		}
		p.SetHasBlockCollision(true)
		p.SetSilent(false)
		p.CheckGroundState(0, 0, 0, 0, 0, 0)
	}
}

// SetGamemode is a port of Player::setGamemode.
func (p *Player) SetGamemode(gm GameMode) bool {
	if p.gamemode == gm {
		return false
	}

	ev := playerevent.NewPlayerGameModeChangeEvent(p, gm)
	event.Call(ev)
	if ev.IsCancelled() {
		return false
	}

	p.internalSetGameMode(gm)

	if p.IsSpectator() {
		p.DespawnFromAll()
	} else {
		p.SpawnToAll()
	}

	p.GetNetworkSession().SyncGameMode(p.gamemode, false)
	return true
}

// IsSurvival reports whether the player is in survival (or adventure, which shares survival's
// behaviour) mode; IsSurvivalLiteral is Player::isSurvival(true).
func (p *Player) IsSurvival() bool {
	return p.gamemode == GameModeSurvival || p.gamemode == GameModeAdventure
}

func (p *Player) IsSurvivalLiteral() bool { return p.gamemode == GameModeSurvival }

// IsCreative reports whether the player is in creative (or spectator) mode; IsCreativeLiteral is
// Player::isCreative(true).
func (p *Player) IsCreative() bool {
	return p.gamemode == GameModeCreative || p.gamemode == GameModeSpectator
}

func (p *Player) IsCreativeLiteral() bool { return p.gamemode == GameModeCreative }

// IsAdventure reports whether the player is in adventure (or spectator) mode;
// IsAdventureLiteral is Player::isAdventure(true).
func (p *Player) IsAdventure() bool {
	return p.gamemode == GameModeAdventure || p.gamemode == GameModeSpectator
}

func (p *Player) IsAdventureLiteral() bool { return p.gamemode == GameModeAdventure }

func (p *Player) IsSpectator() bool { return p.gamemode == GameModeSpectator }

func (p *Player) SetSneakPressed(sneakPressed bool) { p.sneakPressed = sneakPressed }

// IsSneakPressed returns whether the player is pressing the sneak key. The player may still be
// sneaking even if this is false due to gameplay mechanics.
func (p *Player) IsSneakPressed() bool { return p.sneakPressed }

// HasFiniteResources is a port of Player::hasFiniteResources.
func (p *Player) HasFiniteResources() bool { return p.gamemode != GameModeCreative }

// GetInAirTicks is a port of Player::getInAirTicks.
func (p *Player) GetInAirTicks() int { return p.inAirTicks }

// IsSpawned is Player::$spawned.
func (p *Player) IsSpawned() bool { return p.spawned }

// GetLogger returns the player's logger.
func (p *Player) GetLogger() log.Logger { return p.logger }

// CanSaveWithChunk is a port of Player::canSaveWithChunk.
func (p *Player) CanSaveWithChunk() bool { return false }

// NeverSavedWithChunk marks Player as a NeverSavedWithChunkEntity.
func (p *Player) NeverSavedWithChunk() {}

// GetSaveData is a port of Player::getSaveData.
func (p *Player) GetSaveData() *nbt.CompoundTag {
	tag := p.SaveNBT()

	tag.SetString(TagLastKnownXUID, nbt.StringTag(p.xuid))

	if loc := p.GetLocation(); loc.IsValid() {
		tag.SetString(TagLevel, nbt.StringTag(p.GetWorld().GetFolderName()))
	}

	if p.HasValidCustomSpawn() {
		spawn := p.GetSpawn()
		tag.SetString(tagSpawnWorld, nbt.StringTag(spawn.World.GetFolderName()))
		tag.SetInt(tagSpawnX, nbt.IntTag(spawn.FloorX()))
		tag.SetInt(tagSpawnY, nbt.IntTag(spawn.FloorY()))
		tag.SetInt(tagSpawnZ, nbt.IntTag(spawn.FloorZ()))
	}

	if p.deathPosition.IsValid() {
		tag.SetString(tagDeathWorld, nbt.StringTag(p.deathPosition.World.GetFolderName()))
		tag.SetInt(tagDeathX, nbt.IntTag(p.deathPosition.FloorX()))
		tag.SetInt(tagDeathY, nbt.IntTag(p.deathPosition.FloorY()))
		tag.SetInt(tagDeathZ, nbt.IntTag(p.deathPosition.FloorZ()))
	}

	tag.SetInt(tagGameMode, nbt.IntTag(p.gamemode))
	tag.SetLong(TagFirstPlayed, nbt.LongTag(p.firstPlayed))
	tag.SetLong(TagLastPlayed, nbt.LongTag(time.Now().UnixMilli()))

	return tag
}

// Save is a port of Player::save: handles player data saving.
func (p *Player) Save() {
	p.server.SaveOfflinePlayerData(p.username, p.GetSaveData())
}

// Close is Entity::close. It's declared here because *permission.Permissible (PHP's
// PermissibleDelegateTrait) also has a Close, which would otherwise shadow the entity's.
func (p *Player) Close() { p.Human.Close() }
