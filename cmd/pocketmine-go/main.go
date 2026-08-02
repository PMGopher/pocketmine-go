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
// inventory, and every other gameplay packet beyond spawning and seeing terrain aren't wired up
// yet - see handleConn's read loop.
package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"

	"pocketmine-go/pocketmine"
	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/log"
	"pocketmine-go/pocketmine/network/mcpe/convert"
	"pocketmine-go/pocketmine/network/mcpe/serializer"
	"pocketmine-go/pocketmine/world"
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
	seed := flag.Int("seed", 0, "world seed")
	flag.Parse()

	logger := log.NewSimpleLogger()
	log.SetGlobal(logger)

	translator := convert.NewBlockTranslator()
	gen := generator.NewNormal(*seed)
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
	})

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

	spawnY := computeSpawnY(w)
	logger.Info(fmt.Sprintf("world seed %d, spawn height %d", *seed, spawnY))

	go acceptLoop(listener, w, int64(*seed), spawnY, logger)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig

	logger.Info("Shutting down...")
}

// computeSpawnY finds the real ground level at the world origin (0,0) - unlike Flat's fixed-height
// layer stack, Normal's terrain height varies by position, so this can't be a constant.
func computeSpawnY(w *world.World) int32 {
	chunk := w.GetOrLoadChunk(0, 0)
	if h, ok := chunk.GetHighestBlockAt(0, 0); ok {
		return int32(h) + 1
	}
	return 64
}

func acceptLoop(listener *minecraft.Listener, w *world.World, seed int64, spawnY int32, logger log.Logger) {
	for {
		c, err := listener.Accept()
		if err != nil {
			// Accept only errors once the listener has been closed.
			return
		}
		go handleConn(c.(*minecraft.Conn), listener, w, seed, spawnY, logger)
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
func handleConn(conn *minecraft.Conn, listener *minecraft.Listener, w *world.World, seed int64, spawnY int32, logger log.Logger) {
	defer conn.Close()
	defer listener.Disconnect(conn, "server closed")

	name := conn.IdentityData().DisplayName
	logger.Info(fmt.Sprintf("%s connecting from %s", name, conn.RemoteAddr()))

	data := minecraft.GameData{
		WorldName:       pocketmine.Name,
		WorldSeed:       seed,
		Difficulty:      2, // normal
		EntityUniqueID:  1,
		EntityRuntimeID: 1,
		PlayerGameMode:  0, // survival
		PlayerPosition:  mgl32.Vec3{0.5, float32(spawnY), 0.5},
		WorldSpawn:      protocol.BlockPos{0, spawnY, 0},
		WorldGameMode:   0,
		Time:            6000,
		GameRules:       []protocol.GameRule{{Name: "showcoordinates", Value: true}},
	}
	if err := conn.StartGame(data); err != nil {
		logger.Warning(fmt.Sprintf("%s failed to start game: %v", name, err))
		return
	}
	logger.Info(fmt.Sprintf("%s spawned, sending terrain...", name))

	if err := sendSpawnChunks(conn, w, spawnY); err != nil {
		logger.Warning(fmt.Sprintf("%s: failed to send terrain: %v", name, err))
		return
	}
	logger.Info(fmt.Sprintf("%s: terrain sent (%d chunks)", name, (2*spawnChunkRadius+1)*(2*spawnChunkRadius+1)))

	for {
		pk, err := conn.ReadPacket()
		if err != nil {
			logger.Info(fmt.Sprintf("%s disconnected", name))
			return
		}
		switch pk.(type) {
		case *packet.Text:
			// Chat isn't wired to anything yet - just proves packets round-trip both ways.
		}
	}
}

// sendSpawnChunks sends every chunk in a fixed square (see spawnChunkRadius) around the world
// origin, then a NetworkChunkPublisherUpdate telling the client that area is loaded - the same two
// pieces PocketMine-MP's own chunk-sending path needs, minus real per-player view-distance-driven
// streaming as chunks come in and out of range (this port sends one fixed area to every player
// today, regardless of where they go afterwards).
func sendSpawnChunks(conn *minecraft.Conn, w *world.World, spawnY int32) error {
	for x := -spawnChunkRadius; x <= spawnChunkRadius; x++ {
		for z := -spawnChunkRadius; z <= spawnChunkRadius; z++ {
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
		Position: protocol.BlockPos{0, spawnY, 0},
		Radius:   radiusBlocks,
	})
}
