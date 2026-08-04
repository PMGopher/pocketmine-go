// Command pocketmine-go is the entry point for this port: a standalone binary that opens a real
// Minecraft Bedrock listener, lets a client connect and spawn, and sends it real generated terrain
// (Normal - noise-shaped, biome-varied hills/oceans/deserts/... - the same generator real
// PocketMine-MP defaults to) to stand on.
//
// The transport (RakNet) and the entire Bedrock game protocol on top of it (login, the encryption
// handshake, resource pack negotiation, ...) are both provided by
// github.com/sandertv/gophertunnel - the same library (by the same author as go-raknet) that
// Dragonfly is built on. Hand-porting PocketMine-MP's network/mcpe/protocol package from PHP would
// be its own enormous, mostly mechanical undertaking with no relation to PocketMine-MP's actual
// game logic, which is what the rest of this port focuses on - gophertunnel gets a client all the
// way to spawning in far less time than reimplementing the wire format from scratch would.
//
// This is still an early milestone, not a playable server: block placing and real inventory
// interaction (ItemStackRequest handling) aren't wired up yet - see handleConn's read loop.
package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"

	"pocketmine-go/pocketmine"
	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/data/bedrock"
	"pocketmine-go/pocketmine/log"
	pmmath "pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/network/mcpe/convert"
	"pocketmine-go/pocketmine/network/mcpe/serializer"
	"pocketmine-go/pocketmine/player"
	"pocketmine-go/pocketmine/world"
	worldio "pocketmine-go/pocketmine/world/format/io"
	"pocketmine-go/pocketmine/world/generator"
)

// spawnChunkRadius is both the default view distance (in chunks) given to every connecting player
// (see streamChunksToPlayer/player.Player.SetViewDistance - real per-player view distance is now
// genuinely streamed as the player moves, not just sent once) and the radius kept permanently
// loaded around the world spawn regardless of players (see registerSpawnAreaAsPermanentlyLoaded),
// matching PocketMine-MP's own separate "spawn radius" concept in spirit.
const spawnChunkRadius = 4

func main() {
	port := flag.Int("port", 19132, "UDP port to listen on")
	motd := flag.String("motd", pocketmine.Name, "message of the day shown in the server list")
	maxPlayers := flag.Int("max-players", 20, "player count advertised in the server list")
	seed := flag.Int("seed", 0, "world seed (only used the first time a world is created)")
	worldDir := flag.String("world-dir", "world", "directory to save/load the world's LevelDB data in")
	flag.Parse()

	logger := log.NewSimpleLogger()
	log.SetGlobal(logger)

	// level.dat (see pocketmine/world/format/io) is what makes -seed only matter "the first time a
	// world is created" (this flag's own help text, above): once a world already has one, its
	// stored seed always wins - a restart with a different -seed value must never silently
	// regenerate different terrain under an existing save.
	levelDatPath := filepath.Join(*worldDir, "level.dat")
	_, levelDatErr := os.Stat(levelDatPath)
	levelDatExists := levelDatErr == nil

	resolvedSeed := int64(*seed)
	var wd *worldio.WorldData
	if levelDatExists {
		loaded, err := worldio.LoadWorldData(*worldDir)
		if err != nil {
			logger.Critical(fmt.Sprintf("failed to load %q: %v", levelDatPath, err))
			os.Exit(1)
		}
		wd = loaded
		resolvedSeed = wd.GetSeed()
		if int64(*seed) != 0 && int64(*seed) != resolvedSeed {
			logger.Notice(fmt.Sprintf("ignoring -seed=%d: world %q was already created with seed %d", *seed, *worldDir, resolvedSeed))
		}
	}

	translator := convert.NewBlockTranslator()
	gen := generator.NewNormal(int(resolvedSeed))
	w := world.New(gen, translator, []block.Behavior{
		block.VanillaAir(),
		block.VanillaBedrock(),
		block.VanillaStone(),
		block.VanillaDirt(),
		block.VanillaGrass(),
		block.VanillaGravel(),
		block.VanillaCoalOre(),
		block.VanillaIronOre(),
		block.VanillaRedstoneOre(),
		block.VanillaLapisLazuliOre(),
		block.VanillaGoldOre(),
		block.VanillaDiamondOre(),
		block.VanillaEmeraldOre(),
		block.VanillaWater(),
		block.VanillaSand(),
		block.VanillaSandstone(),
		block.VanillaSnowLayer(),
		block.VanillaTallGrass(),
		block.VanillaOakLog(),
		block.VanillaOakLeaves(),
		block.VanillaSpruceLog(),
		block.VanillaSpruceLeaves(),
		block.VanillaBirchLog(),
		block.VanillaBirchLeaves(),
	})

	if err := w.OpenProvider(*worldDir); err != nil {
		logger.Critical(fmt.Sprintf("failed to open world at %q: %v", *worldDir, err))
		os.Exit(1)
	}
	defer func() {
		logger.Info("Saving world...")
		if err := w.Close(); err != nil {
			logger.Warning(fmt.Sprintf("failed to save world: %v", err))
		}
		if wd != nil {
			wd.SetTime(w.GetTime())
			if err := wd.Save(*worldDir); err != nil {
				logger.Warning(fmt.Sprintf("failed to save %q: %v", levelDatPath, err))
			}
		}
	}()

	cfg := minecraft.ListenConfig{
		StatusProvider: minecraft.NewStatusProvider(*motd, pocketmine.Name),
		// Xbox Live authentication needs a real Microsoft account handshake and outbound HTTPS
		// access neither of which this early milestone needs - disabling it is the same tradeoff
		// PocketMine-MP's own `xbox-auth: false` setting makes for offline/LAN play.
		AuthenticationDisabled: true,
		MaximumPlayers:         *maxPlayers,
	}

	addr := ":" + strconv.Itoa(*port)
	listener, err := cfg.Listen("raknet", addr)
	if err != nil {
		logger.Critical(fmt.Sprintf("failed to start listener on %s: %v", addr, err))
		os.Exit(1)
	}
	defer listener.Close()

	logger.Info(fmt.Sprintf("%s %s is listening on %s", pocketmine.Name, pocketmine.Version().GetFullVersion(true), addr))

	spawn := computeSpawn(w)
	logger.Info(fmt.Sprintf("world seed %d, spawn at %d,%d,%d", resolvedSeed, spawn.X, spawn.Y, spawn.Z))
	spawnVec := pmmath.NewVector3(float64(spawn.X), float64(spawn.Y), float64(spawn.Z))
	w.SetSpawnLocation(spawnVec)

	if !levelDatExists {
		generated, err := worldio.GenerateWorldData(*worldDir, pocketmine.Name, resolvedSeed, worldio.GeneratorInfinite, "normal", "", spawnVec)
		if err != nil {
			logger.Warning(fmt.Sprintf("failed to write %q: %v", levelDatPath, err))
		} else {
			wd = generated
		}
	}

	// Players now stream chunks around themselves individually as they move (see
	// streamChunksToPlayer/player.Player.OrderChunks), but the spawn area itself is still kept
	// permanently loaded/ticking server-side regardless of whether any player is nearby - matching
	// real PocketMine-MP's own separate "spawn-radius keeps chunks loaded" behaviour, under one
	// shared loader identity rather than per-connection tokens.
	registerSpawnAreaAsPermanentlyLoaded(w, spawn)

	go runTickLoop(w)

	reg := newRegistry(logger)
	go acceptLoop(listener, w, reg, resolvedSeed, spawn, logger)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig

	logger.Info("Shutting down...")
}

// spawnMaxSearchRadius bounds computeSpawn's outward search for dry land - generous enough to
// escape all but exceptionally large open-ocean regions (see Normal's own doc comment on why a
// single ocean biome can occasionally span a wide area).
const spawnMaxSearchRadius = 128

type spawnPoint struct{ X, Y, Z int32 }

// computeSpawn finds a safe spawn point near the world origin: the first column, searching
// outward in a square spiral, whose surface isn't a liquid (matching the intent, if not the exact
// algorithm, of World::getSafeSpawn - real PocketMine-MP's search is chunk-order/border-aware in
// ways this port's single in-memory World doesn't need). Falls back to (0, 64, 0) if nothing
// suitable turns up within spawnMaxSearchRadius blocks, which should only happen deep inside an
// unusually large ocean.
func computeSpawn(w *world.World) spawnPoint {
	for radius := 0; radius <= spawnMaxSearchRadius; radius++ {
		for x := -radius; x <= radius; x++ {
			for z := -radius; z <= radius; z++ {
				if radius > 0 && x != -radius && x != radius && z != -radius && z != radius {
					// Only the ring at exactly this radius is new - smaller radii were already
					// checked on earlier iterations.
					continue
				}
				chunk := w.GetOrLoadChunk(x>>4, z>>4)
				h, ok := chunk.GetHighestBlockAt(x&0xf, z&0xf)
				if !ok {
					continue
				}
				if top := w.GetBlockAt(x, h, z); !block.IsLiquid(top) {
					return spawnPoint{int32(x), int32(h) + 1, int32(z)}
				}
			}
		}
	}
	return spawnPoint{0, 64, 0}
}

// spawnAreaLoader is the shared identity token registered with every chunk in the spawn area (see
// registerSpawnAreaAsPermanentlyLoaded) - World.RegisterChunkLoader/RegisterTickingChunk only need
// something comparable for identity (real PocketMine-MP's ChunkLoader/ChunkTicker are themselves
// empty marker interfaces - see World.RegisterChunkLoader's own doc comment), so an unexported
// empty struct works exactly as well as a real object would.
type spawnAreaLoader struct{}

// registerSpawnAreaAsPermanentlyLoaded generates, registers as loaded (so it's never eligible for
// World.UnregisterChunkLoader-driven unloading) and registers for random ticking every chunk in a
// fixed spawnChunkRadius area around spawn - independent of streamChunksToPlayer's own per-player
// loader registrations for the same chunks.
func registerSpawnAreaAsPermanentlyLoaded(w *world.World, spawn spawnPoint) {
	loader := &spawnAreaLoader{}
	centerX, centerZ := int(spawn.X)>>4, int(spawn.Z)>>4
	for x := centerX - spawnChunkRadius; x <= centerX+spawnChunkRadius; x++ {
		for z := centerZ - spawnChunkRadius; z <= centerZ+spawnChunkRadius; z++ {
			w.GetOrLoadChunk(x, z)
			w.RegisterChunkLoader(loader, x, z)
			w.RegisterTickingChunk(loader, x, z)
		}
	}
}

// ticksPerSecond matches real PocketMine-MP/Minecraft's fixed 20 TPS.
const ticksPerSecond = 20

// runTickLoop drives World.DoTick once per game tick for as long as the process runs - this port's
// equivalent of real PocketMine-MP's Server's own tick timer calling World::doTick on every loaded
// world. It's never explicitly stopped: main's own shutdown path (the signal wait in main()) exits
// the whole process right after, which is enough to stop this goroutine too - the same "no graceful
// per-goroutine shutdown needed" reasoning acceptLoop's own goroutine already relies on.
func runTickLoop(w *world.World) {
	ticker := time.NewTicker(time.Second / ticksPerSecond)
	defer ticker.Stop()

	var currentTick int64
	for range ticker.C {
		currentTick++
		w.DoTick(currentTick)
	}
}

func acceptLoop(listener *minecraft.Listener, w *world.World, reg *registry, seed int64, spawn spawnPoint, logger log.Logger) {
	for {
		c, err := listener.Accept()
		if err != nil {
			// Accept only errors once the listener has been closed.
			return
		}
		go handleConn(c.(*minecraft.Conn), listener, w, reg, seed, spawn, logger)
	}
}

// handleConn drives one player's connection for as long as it stays open. By the time Accept
// returns a *minecraft.Conn, gophertunnel has already completed the entire login handshake
// (encryption, resource pack negotiation) internally - StartGame is the first thing this port
// itself is responsible for.
//
// GameData.Items is now real - built from the vendored required_item_list.json (see
// pocketmine/data/bedrock.ItemTypes, the item-table counterpart of BlockStates) rather than left
// empty. This is the client's whole vocabulary of known item names/network IDs - without it every
// item renders as an unknown/blank icon regardless of what this port's own item registry supports.
// itemTable builds the StartGame/ItemRegistry item table from the real vendored Bedrock item list
// (see itemTableCache's own doc comment on why this is computed once and reused).
func itemTable() []protocol.ItemEntry {
	itemTableOnce.Do(func() {
		types := bedrock.ItemTypes()
		itemTableCache = make([]protocol.ItemEntry, len(types))
		for i, t := range types {
			itemTableCache[i] = protocol.ItemEntry{
				Name:           t.Name,
				RuntimeID:      int16(t.RuntimeID),
				ComponentBased: t.ComponentBased,
				Version:        t.Version,
				Data:           t.Data,
			}
		}
	})
	return itemTableCache
}

// itemTableOnce/itemTableCache memoise itemTable() - every connecting player gets the exact same
// item table (it's a fixed protocol-version-wide vocabulary, not per-player state), so there's no
// reason to rebuild the ~1900-entry slice (with its component NBT) on every single connection.
var (
	itemTableOnce  sync.Once
	itemTableCache []protocol.ItemEntry
)

// survivalAbilities builds a real UpdateAbilities packet with survival-appropriate defaults (can
// build/mine/interact/attack, no flying/noclip/invulnerability) - mirrors what a fresh survival
// PocketMine-MP player gets. Sent once right after StartGame, and again any time the client sends
// RequestAbility (see handleConn's read loop) - the client predicts an ability change locally the
// moment it sends the request, and without an authoritative reply the prediction never gets
// corrected. This is exactly what was happening with Flying specifically: swimming to the surface
// sends the same "double-jump" input gesture Creative mode uses to toggle flight, the client
// requested AbilityFlying=true, and since nothing ever told it no, it kept flying indefinitely.
// Re-sending this exact packet (still with Flying/MayFly/NoClip left false) is the correction.
func survivalAbilities(entityUniqueID int64) *packet.UpdateAbilities {
	abilities := uint32(protocol.AbilityBuild | protocol.AbilityMine | protocol.AbilityDoorsAndSwitches |
		protocol.AbilityOpenContainers | protocol.AbilityAttackPlayers | protocol.AbilityAttackMobs)
	return &packet.UpdateAbilities{AbilityData: protocol.AbilityData{
		EntityUniqueID:     entityUniqueID,
		PlayerPermissions:  packet.PermissionLevelMember,
		CommandPermissions: protocol.CommandPermissionLevelAny,
		Layers: []protocol.AbilityLayer{{
			Type:      protocol.AbilityLayerTypeBase,
			Abilities: protocol.AbilityCount - 1,
			Values:    abilities,
			WalkSpeed: protocol.AbilityBaseWalkSpeed,
		}},
	}}
}

func handleConn(conn *minecraft.Conn, listener *minecraft.Listener, w *world.World, reg *registry, seed int64, spawn spawnPoint, logger log.Logger) {
	defer conn.Close()
	defer listener.Disconnect(conn, "server closed")

	name := conn.IdentityData().DisplayName
	logger.Info(fmt.Sprintf("%s connecting from %s", name, conn.RemoteAddr()))

	spawnVec := mgl32.Vec3{float32(spawn.X) + 0.5, float32(spawn.Y), float32(spawn.Z) + 0.5}
	sess, err := newSession(conn, w, spawnVec)
	if err != nil {
		logger.Warning(fmt.Sprintf("%s: failed to build session: %v", name, err))
		return
	}

	data := minecraft.GameData{
		WorldName:       pocketmine.Name,
		WorldSeed:       seed,
		Difficulty:      2, // normal
		EntityUniqueID:  sess.entityUniqueID,
		EntityRuntimeID: sess.entityRuntimeID,
		PlayerGameMode:  0, // survival
		PlayerPosition:  spawnVec,
		WorldSpawn:      protocol.BlockPos{spawn.X, spawn.Y, spawn.Z},
		WorldGameMode:   0,
		Time:            6000,
		GameRules:       []protocol.GameRule{{Name: "showcoordinates", Value: true}},
		Items:           itemTable(),
	}
	if err := conn.StartGame(data); err != nil {
		logger.Warning(fmt.Sprintf("%s failed to start game: %v", name, err))
		return
	}

	// Real Bedrock servers always follow StartGame with UpdateAbilities - without it the client has
	// no ability layer at all for its local player, which (empirically) can leave it refusing to
	// process movement input.
	if err := conn.WritePacket(survivalAbilities(data.EntityUniqueID)); err != nil {
		logger.Warning(fmt.Sprintf("%s: failed to send abilities: %v", name, err))
		return
	}
	logger.Info(fmt.Sprintf("%s spawned, sending terrain...", name))

	sess.player.SetViewDistance(spawnChunkRadius)
	sent, err := streamChunksToPlayer(conn, sess.player)
	if err != nil {
		logger.Warning(fmt.Sprintf("%s: failed to send terrain: %v", name, err))
		return
	}
	logger.Info(fmt.Sprintf("%s: terrain sent (%d chunks)", name, sent))

	// Join makes every already-connected player visible to this one and vice versa (PlayerList +
	// AddPlayer both ways) - see registry's doc comment. Leave (on disconnect, below) reverses it.
	reg.Join(sess)
	defer reg.Leave(sess)
	logger.Info(fmt.Sprintf("%s: joined (%d player(s) online)", name, reg.Count()))

	for {
		pk, err := conn.ReadPacket()
		if err != nil {
			logger.Info(fmt.Sprintf("%s disconnected", name))
			return
		}
		switch input := pk.(type) {
		case *packet.Text:
			// Chat isn't broadcast to anyone yet - just proves packets round-trip both ways.
		case *packet.PlayerAuthInput:
			// The client reports its own predicted position/rotation every tick once
			// server-authoritative movement is active (see PlayerAuthInput's own doc comment - this
			// is now the only movement path modern Bedrock versions speak, MovePlayer/client-
			// authoritative movement no longer exists in this protocol version). This port has no
			// real physics/collision yet, so the client's report is simply trusted as-is - the same
			// "no correction unless there's a pending teleport" approach real Bedrock servers use
			// for ordinary movement (see e.g. Dragonfly's PlayerAuthInputHandler) - then relayed to
			// every other connected player so they see this player move.
			sess.SetPositionAndRotation(input.Position, input.Pitch, input.Yaw, input.HeadYaw)
			reg.BroadcastMove(sess)
			// InputFlagVerticalCollision is the client's own report of touching something
			// vertically (ground or ceiling) this tick - the real on-ground signal this port has,
			// since (like position/rotation above) there's no server-side physics to derive it from
			// independently. Drives real fall-damage tracking (see Player.TrackFallState's own doc
			// comment).
			sess.player.TrackFallState(float64(input.Position[1]), input.InputData.Load(packet.InputFlagVerticalCollision))
			if _, err := streamChunksToPlayer(conn, sess.player); err != nil {
				logger.Warning(fmt.Sprintf("%s: failed to stream terrain: %v", name, err))
				return
			}
			sess.player.UpdateBreakingBlock(bareHandItem{})
			handleBlockActions(conn, sess.player, input.BlockActions, logger, name)
		case *packet.InventoryTransaction:
			handleInventoryTransaction(sess, reg, input, logger, name)
		case *packet.RequestAbility:
			// This port doesn't support granting any client-requested ability yet (no flying, no
			// noclip - see survivalAbilities' own doc comment on why simply not replying isn't an
			// option: the client predicts the request succeeded until told otherwise). Re-asserting
			// the same survival ability set denies every request uniformly.
			if err := conn.WritePacket(survivalAbilities(sess.entityUniqueID)); err != nil {
				logger.Warning(fmt.Sprintf("%s: failed to re-send abilities: %v", name, err))
				return
			}
		}
	}
}

// handleInventoryTransaction is a port of the entity-attack half of PlayerAuthInput/
// InventoryTransaction handling onto Player's own real AttackEntity - real PvP (damage, knockback,
// hit sound, arm-swing/hurt animations). Only UseItemOnEntityTransactionData with
// UseItemOnEntityActionAttack is handled - Interact (right-click) isn't wired to anything yet (no
// entity-interact use case exists in this port - villager trading, boat/mount riding, etc.).
func handleInventoryTransaction(sess *session, reg *registry, pk *packet.InventoryTransaction, logger log.Logger, name string) {
	data, ok := pk.TransactionData.(*protocol.UseItemOnEntityTransactionData)
	if !ok || data.ActionType != protocol.UseItemOnEntityActionAttack {
		return
	}

	target, ok := reg.Get(data.TargetEntityRuntimeID)
	if !ok {
		return
	}

	if sess.player.AttackEntity(target.player, bareHandItem{}) {
		logger.Info(fmt.Sprintf("%s attacked %s", name, target.name))
	}
}

// bareHandItem is a stand-in block.Item for "whatever the player is currently holding" - this
// port's inventory isn't wired to a "selected hotbar slot" concept yet (no ItemStackRequest
// handling exists - see this file's own package doc comment), so every break/attack action is
// computed as if the player were holding nothing (efficiency 1.0, no tool type), matching a bare
// hand exactly. A real held-item lookup replaces this once hotbar selection exists.
type bareHandItem struct{}

func (bareHandItem) GetTypeId() int                                        { return 0 }
func (bareHandItem) GetBlockToolType() block.ToolType                      { return block.ToolTypeNone }
func (bareHandItem) GetBlockToolHarvestLevel() int                         { return 0 }
func (bareHandItem) GetMiningEfficiency(isCompatibleToolType bool) float64 { return 1 }
func (bareHandItem) Pop()                                                  {}
func (bareHandItem) IsNull() bool                                          { return false }
func (bareHandItem) GetCustomName() string                                 { return "" }
func (bareHandItem) GetCount() int                                         { return 1 }
func (bareHandItem) SetCount(count int)                                    {}

// handleBlockActions is a port of the block-breaking half of PlayerAuthInput handling (the
// placing half isn't wired up yet - that goes through ItemStackRequest, a separate, inventory-
// shaped undertaking) onto Player's own real AttackBlock/ContinueBreakBlock/StopBreakBlock/
// BreakBlock - a genuine SurvivalBlockBreakHandler now drives the break-time state machine (see
// player.SurvivalBlockBreakHandler's own doc comment for what it still doesn't model: haste/mining
// fatigue/aqua affinity, and the network broadcasts to viewers).
func handleBlockActions(conn *minecraft.Conn, p *player.Player, actions []protocol.PlayerBlockAction, logger log.Logger, name string) {
	for _, action := range actions {
		pos := action.BlockPos
		vec := pmmath.NewVector3(float64(pos[0]), float64(pos[1]), float64(pos[2]))
		face := pmmath.Facing(action.Face)

		switch action.Action {
		case protocol.PlayerActionStartBreak:
			p.AttackBlock(vec, face, bareHandItem{})
		case protocol.PlayerActionContinueDestroyBlock:
			p.ContinueBreakBlock(vec, face)
		case protocol.PlayerActionAbortBreak:
			p.StopBreakBlock(vec)
		case protocol.PlayerActionPredictDestroyBlock, protocol.PlayerActionStopBreak:
			if !p.BreakBlock(vec) {
				continue
			}
			w := p.GetWorld()
			airNetworkID := uint32(w.Translator().InternalIDToNetworkID(block.VanillaAir()))
			if err := conn.WritePacket(&packet.UpdateBlock{
				Position:          pos,
				NewBlockRuntimeID: airNetworkID,
				Flags:             packet.BlockUpdateNetwork,
			}); err != nil {
				logger.Warning(fmt.Sprintf("%s: failed to confirm block break at %v: %v", name, pos, err))
			}
		}
	}
}

// streamChunksToPlayer is a port of the network-sending half of Player::requestChunks: drives p's
// own real OrderChunks/RequestChunks (see player.Player's own doc comment on the real per-player
// view-distance-driven chunk streaming this replaces the old fixed-area broadcast with), sends a
// LevelChunk packet for every newly-ready chunk, marks each one sent, and syncs the client's view
// area center point - callers call this both once at spawn and again on every PlayerAuthInput, so
// chunks stream in as the player moves instead of only ever covering one fixed area. Returns how
// many chunks were sent this call.
func streamChunksToPlayer(conn *minecraft.Conn, p *player.Player) (int, error) {
	p.OrderChunks()
	ready := p.RequestChunks()
	if len(ready) == 0 {
		return 0, nil
	}

	w := p.GetWorld()
	for _, c := range ready {
		chunk, ok := w.GetChunk(c[0], c[1])
		if !ok {
			continue
		}
		payload := serializer.SerializeFullChunk(chunk, w.Translator())
		pk := &packet.LevelChunk{
			Position:      protocol.ChunkPos{int32(c[0]), int32(c[1])},
			SubChunkCount: uint32(serializer.GetSubChunkCount(chunk)),
			RawPayload:    payload,
		}
		if err := conn.WritePacket(pk); err != nil {
			return 0, err
		}
		p.MarkChunkSent(c[0], c[1])
	}

	pos := p.GetPosition()
	radiusBlocks := uint32(p.GetViewDistance()) << 4
	if err := conn.WritePacket(&packet.NetworkChunkPublisherUpdate{
		Position: protocol.BlockPos{int32(pos.X), int32(pos.Y), int32(pos.Z)},
		Radius:   radiusBlocks,
	}); err != nil {
		return 0, err
	}

	return len(ready), nil
}
