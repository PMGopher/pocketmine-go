// Command pocketmine-go is the first real entry point for this port: a standalone binary that
// opens a RakNet listener (via github.com/sandertv/go-raknet) on the standard Bedrock port and
// answers unconnected pings, so the server shows up in a Minecraft client's server list with a
// MOTD and player count.
//
// This is deliberately NOT a full Minecraft Bedrock server yet. RakNet only gets a client to the
// point of a raw, reliable connection - the actual Bedrock game protocol on top of it (login,
// encryption handshake, resource packs, chunk/entity data, ...) isn't implemented at all in this
// port yet. A client that tries to actually join will connect at the RakNet level (see the log
// line for each accepted connection) and then time out, since no login response is ever sent.
package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/sandertv/go-raknet"

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

	addr := ":" + strconv.Itoa(*port)
	listener, err := raknet.Listen(addr)
	if err != nil {
		logger.Critical(fmt.Sprintf("failed to start RakNet listener on %s: %v", addr, err))
		os.Exit(1)
	}
	defer listener.Close()

	listener.PongData([]byte(buildPongData(*motd, *port, 0, *maxPlayers, listener.ID())))

	logger.Info(fmt.Sprintf("%s %s is listening on %s (RakNet only - no Bedrock game protocol yet)", pocketmine.Name, pocketmine.Version().GetFullVersion(true), addr))

	go acceptLoop(listener, logger)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig

	logger.Info("Shutting down...")
}

// acceptLoop accepts RakNet connections and logs each one. It doesn't (and can't yet) speak the
// Bedrock game protocol on top of the connection, so it just proves the transport layer is real
// and working, then closes the connection.
func acceptLoop(listener *raknet.Listener, logger log.Logger) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			// Accept only errors once the listener has been closed.
			return
		}
		go handleConn(conn, logger)
	}
}

func handleConn(conn net.Conn, logger log.Logger) {
	defer conn.Close()
	logger.Info(fmt.Sprintf("RakNet connection accepted from %s (no Bedrock login handling yet - it will time out)", conn.RemoteAddr()))

	buf := make([]byte, 1500)
	n, err := conn.Read(buf)
	if err != nil {
		return
	}
	logger.Debug(fmt.Sprintf("received %d bytes from %s (packet ID 0x%02x) - discarded, Bedrock protocol not implemented", n, conn.RemoteAddr(), buf[0]))
}

// buildPongData builds the MCPE unconnected-pong payload Minecraft Bedrock clients parse to show
// this server in their server list: a semicolon-separated string of
// "MCPE;<motd>;<protocol version>;<version name>;<player count>;<max players>;<server GUID>;
// <motd2>;<gamemode>;<gamemode numeric>;<port ipv4>;<port ipv6>;".
//
// The protocol version/version name fields are placeholders (this port doesn't speak the real
// Bedrock protocol yet, so no actual version negotiation happens) - real clients will still show
// the server in their list, but will fail to join past the RakNet handshake.
func buildPongData(motd string, port, playerCount, maxPlayers int, serverGUID int64) string {
	const (
		protocolVersion = 0
		versionName     = "0.0.0"
		gamemode        = "Survival"
		gamemodeNumeric = 1
	)
	return fmt.Sprintf(
		"MCPE;%s;%d;%s;%d;%d;%d;%s;%s;%d;%d;%d;",
		motd, protocolVersion, versionName, playerCount, maxPlayers, serverGUID,
		motd, gamemode, gamemodeNumeric, port, port,
	)
}
