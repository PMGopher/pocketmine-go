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
// This is still an early milestone, not a playable server: block interaction (breaking/placing),
// inventory, and every other gameplay packet beyond spawning, moving around and seeing terrain
// aren't wired up yet - see handleConn's read loop.
package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"

	"pocketmine-go/pocketmine"
	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/log"
	pmmath "pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/network/mcpe/convert"
	"pocketmine-go/pocketmine/network/mcpe/serializer"
	"pocketmine-go/pocketmine/world"
	worldio "pocketmine-go/pocketmine/world/format/io"
	"pocketmine-go/pocketmine/world/generator"
)

// spawnChunkRadius is how many chunks in every direction around (0,0) get sent to a joining
// player, matching PocketMine-MP's own "spawn radius" concept (chunks always kept loaded/sent
// around the world spawn) in spirit, though real per-player view-distance-driven chunk streaming
// isn't ported - every player currently gets this same fixed area regardless of movement.
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

	// This port sends every player the same fixed spawnChunkRadius area (see sendSpawnChunks' own
	// doc comment on the lack of real per-player view-distance streaming) rather than loading
	// chunks per-player as they move - so, matching that, the spawn area itself is registered as a
	// single permanently-loaded/ticking region for the server's whole lifetime (akin to real
	// PocketMine-MP's own "spawn-radius keeps chunks loaded" behaviour), under one shared loader
	// identity rather than per-connection tokens.
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
// World.UnregisterChunkLoader-driven unloading) and registers for random ticking every chunk in
// the same fixed area sendSpawnChunks sends to every player.
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
// GameData.Items is left empty: gophertunnel is a protocol library, not a game implementation, so
// it doesn't ship the vanilla item table the way Dragonfly's separate data package does, and this
// port doesn't have one yet either (its own VanillaBlocks/VanillaItems registries only cover the
// handful of types wired up so far - see pocketmine/block/vanilla_blocks.go). A client can still
// complete StartGame/spawn without it, just with incomplete item names/icons until that's filled
// in.
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
	}
	if err := conn.StartGame(data); err != nil {
		logger.Warning(fmt.Sprintf("%s failed to start game: %v", name, err))
		return
	}

	// Real Bedrock servers always follow StartGame with UpdateAbilities - without it the client has
	// no ability layer at all for its local player, which (empirically) can leave it refusing to
	// process movement input. Survival-appropriate defaults (can build/mine/interact/attack, no
	// flying/noclip/invulnerability) mirror what a fresh survival PocketMine-MP player gets.
	abilities := uint32(protocol.AbilityBuild | protocol.AbilityMine | protocol.AbilityDoorsAndSwitches |
		protocol.AbilityOpenContainers | protocol.AbilityAttackPlayers | protocol.AbilityAttackMobs)
	if err := conn.WritePacket(&packet.UpdateAbilities{AbilityData: protocol.AbilityData{
		EntityUniqueID:     data.EntityUniqueID,
		PlayerPermissions:  packet.PermissionLevelMember,
		CommandPermissions: protocol.CommandPermissionLevelAny,
		Layers: []protocol.AbilityLayer{{
			Type:      protocol.AbilityLayerTypeBase,
			Abilities: protocol.AbilityCount - 1,
			Values:    abilities,
			WalkSpeed: protocol.AbilityBaseWalkSpeed,
		}},
	}}); err != nil {
		logger.Warning(fmt.Sprintf("%s: failed to send abilities: %v", name, err))
		return
	}
	logger.Info(fmt.Sprintf("%s spawned, sending terrain...", name))

	if err := sendSpawnChunks(conn, w, spawn); err != nil {
		logger.Warning(fmt.Sprintf("%s: failed to send terrain: %v", name, err))
		return
	}
	logger.Info(fmt.Sprintf("%s: terrain sent (%d chunks)", name, (2*spawnChunkRadius+1)*(2*spawnChunkRadius+1)))

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
			handleBlockActions(conn, w, input.BlockActions, logger, name)
		}
	}
}

// handleBlockActions is a simplified port of the block-breaking half of PlayerAuthInput handling
// (the placing half isn't wired up yet - that goes through ItemStackRequest, a separate,
// inventory-shaped undertaking). Real PocketMine-MP tracks a whole break-time state machine per
// player (BlockBreakInfo's hardness/tool-efficiency/haste calculations, timed across
// StartBreak/ContinueDestroyBlock/AbortBreak), which needs a real Player/inventory/held-item system
// this port doesn't have yet - so for now, any of the "the client believes this block is now
// broken" actions (PredictDestroyBlock, the survival hold-to-break completion; StopBreak, sent for
// creative's instant break) just breaks the block immediately server-side, honestly simplified
// rather than guessed at.
func handleBlockActions(conn *minecraft.Conn, w *world.World, actions []protocol.PlayerBlockAction, logger log.Logger, name string) {
	for _, action := range actions {
		if action.Action != protocol.PlayerActionPredictDestroyBlock && action.Action != protocol.PlayerActionStopBreak {
			continue
		}

		pos := action.BlockPos
		vec := pmmath.NewVector3(float64(pos[0]), float64(pos[1]), float64(pos[2]))
		if !w.UseBreakOn(vec) {
			continue
		}

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

// sendSpawnChunks sends every chunk in a fixed square (see spawnChunkRadius) around the world
// origin, then a NetworkChunkPublisherUpdate telling the client that area is loaded - the same two
// pieces PocketMine-MP's own chunk-sending path needs, minus real per-player view-distance-driven
// streaming as chunks come in and out of range (this port sends one fixed area to every player
// today, regardless of where they go afterwards).
func sendSpawnChunks(conn *minecraft.Conn, w *world.World, spawn spawnPoint) error {
	centerX, centerZ := int(spawn.X)>>4, int(spawn.Z)>>4
	for x := centerX - spawnChunkRadius; x <= centerX+spawnChunkRadius; x++ {
		for z := centerZ - spawnChunkRadius; z <= centerZ+spawnChunkRadius; z++ {
			chunk := w.GetOrLoadChunk(x, z)
			payload := serializer.SerializeFullChunk(chunk, w.Translator())
			pk := &packet.LevelChunk{
				Position:      protocol.ChunkPos{int32(x), int32(z)},
				SubChunkCount: uint32(serializer.GetSubChunkCount(chunk)),
				RawPayload:    payload,
			}
			if err := conn.WritePacket(pk); err != nil {
				return err
			}
		}
	}

	radiusBlocks := uint32(spawnChunkRadius) << 4
	return conn.WritePacket(&packet.NetworkChunkPublisherUpdate{
		Position: protocol.BlockPos{spawn.X, spawn.Y, spawn.Z},
		Radius:   radiusBlocks,
	})
}
