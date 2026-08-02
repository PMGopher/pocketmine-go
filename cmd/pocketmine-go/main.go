// Command pocketmine-go is the entry point for this port: a standalone binary that opens a real
// Minecraft Bedrock listener and lets a client connect and spawn.
//
// The transport (RakNet) and the entire Bedrock game protocol on top of it (login, the encryption
// handshake, resource pack negotiation, ...) are both provided by
// github.com/sandertv/gophertunnel - the same library (by the same author as go-raknet) that
// Dragonfly is built on. Hand-porting PocketMine-MP's network/mcpe/protocol package from PHP would
// be its own enormous, mostly mechanical undertaking with no relation to PocketMine-MP's actual
// game logic, which is what the rest of this port focuses on - gophertunnel gets a client all the
// way to spawning in far less time than reimplementing the wire format from scratch would.
//
// This is still an early milestone, not a playable server: StartGame is sent with no real chunk
// data (see handleConn's doc comment), so a connecting client will get through login and reach
// the point of spawning, but will see an empty/void world since no terrain is sent yet. Real
// chunk data - backed by a real World implementation of the block.World interface the whole
// block package already codes against - is the next milestone.
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
	"pocketmine-go/pocketmine/log"
)

func main() {
	port := flag.Int("port", 19132, "UDP port to listen on")
	motd := flag.String("motd", pocketmine.Name, "message of the day shown in the server list")
	maxPlayers := flag.Int("max-players", 20, "player count advertised in the server list")
	flag.Parse()

	logger := log.NewSimpleLogger()
	log.SetGlobal(logger)

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

	go acceptLoop(listener, logger)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig

	logger.Info("Shutting down...")
}

func acceptLoop(listener *minecraft.Listener, logger log.Logger) {
	for {
		c, err := listener.Accept()
		if err != nil {
			// Accept only errors once the listener has been closed.
			return
		}
		go handleConn(c.(*minecraft.Conn), listener, logger)
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
func handleConn(conn *minecraft.Conn, listener *minecraft.Listener, logger log.Logger) {
	defer conn.Close()
	defer listener.Disconnect(conn, "server closed")

	name := conn.IdentityData().DisplayName
	logger.Info(fmt.Sprintf("%s connecting from %s", name, conn.RemoteAddr()))

	data := minecraft.GameData{
		WorldName:       pocketmine.Name,
		WorldSeed:       0,
		Difficulty:      2, // normal
		EntityUniqueID:  1,
		EntityRuntimeID: 1,
		PlayerGameMode:  0, // survival
		PlayerPosition:  mgl32.Vec3{0.5, 5, 0.5},
		WorldSpawn:      protocol.BlockPos{0, 4, 0},
		WorldGameMode:   0,
		Time:            6000,
		GameRules:       []protocol.GameRule{{Name: "showcoordinates", Value: true}},
	}
	if err := conn.StartGame(data); err != nil {
		logger.Warning(fmt.Sprintf("%s failed to start game: %v", name, err))
		return
	}
	logger.Info(fmt.Sprintf("%s spawned (no terrain yet - see main.go's doc comment)", name))

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
