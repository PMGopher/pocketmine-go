// Command packetproxy is a debugging tool (not part of the PocketMine-MP port): a proxy between a
// real Minecraft client and a server that logs every packet in both directions. It's
// gophertunnel's own proxy example (github.com/sandertv/gophertunnel/main.go) without the Xbox Live
// login, so the server behind it must run with authentication off.
//
// Usage: start the server on another port with authentication off, e.g.
//
//	go run ./cmd/pocketmine-go --server-port=19133 --xbox-auth=false
//
// then start the proxy and join 127.0.0.1:19132 from the game:
//
//	go run ./tools/packetproxy --remote=127.0.0.1:19133
//
// Every packet is written to packets.log; when the connection ends, the last packets are also
// printed. Running the same session against another server (e.g. Dragonfly) shows exactly what
// the client does differently.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/sandertv/gophertunnel/minecraft"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

func main() {
	local := flag.String("listen", "0.0.0.0:19132", "address the game connects to")
	remote := flag.String("remote", "127.0.0.1:19133", "address of the server (authentication off)")
	logPath := flag.String("log", "packets.log", "file every packet is written to")
	flag.Parse()

	logFile, err := os.Create(*logPath)
	if err != nil {
		log.Fatalf("create %s: %v", *logPath, err)
	}
	defer logFile.Close()

	status, err := minecraft.NewForeignStatusProvider(*remote)
	if err != nil {
		log.Fatalf("server %s isn't reachable: %v", *remote, err)
	}
	listener, err := minecraft.ListenConfig{
		StatusProvider:         status,
		AuthenticationDisabled: true,
	}.Listen("raknet", *local)
	if err != nil {
		log.Fatalf("listen on %s: %v", *local, err)
	}
	defer listener.Close()
	log.Printf("Proxy on %s -> %s, logging to %s. Join %s from the game.", *local, *remote, *logPath, *local)

	for {
		c, err := listener.Accept()
		if err != nil {
			return
		}
		go handleConn(c.(*minecraft.Conn), listener, *remote, logFile)
	}
}

// trace writes packets to the log file and keeps the latest ones for the summary.
type trace struct {
	mu     sync.Mutex
	file   *os.File
	start  time.Time
	recent []string
}

func (t *trace) record(direction string, pk packet.Packet) {
	line := fmt.Sprintf("%8.3fs %s %s", time.Since(t.start).Seconds(), direction, describe(pk))
	t.mu.Lock()
	defer t.mu.Unlock()
	fmt.Fprintln(t.file, line)
	t.recent = append(t.recent, line)
	if len(t.recent) > 60 {
		t.recent = t.recent[1:]
	}
}

func (t *trace) note(format string, args ...any) {
	line := fmt.Sprintf("%8.3fs ** %s", time.Since(t.start).Seconds(), fmt.Sprintf(format, args...))
	t.mu.Lock()
	defer t.mu.Unlock()
	fmt.Fprintln(t.file, line)
	t.recent = append(t.recent, line)
}

// describe is the packet type plus the fields that matter for chunk loading.
func describe(pk packet.Packet) string {
	name := strings.TrimPrefix(fmt.Sprintf("%T", pk), "*packet.")
	switch p := pk.(type) {
	case *packet.LevelChunk:
		limit, hasLimit := p.SubChunkLimit.Value()
		return fmt.Sprintf("%s pos=%v dim=%d count=%d limit=%d(%v) cache=%v hashes=%d payload=%d",
			name, p.Position, p.Dimension, p.SubChunkCount, limit, hasLimit, p.CacheEnabled, len(p.BlobHashes), len(p.RawPayload))
	case *packet.SubChunkRequest:
		return fmt.Sprintf("%s dim=%d pos=%v offsets=%d", name, p.Dimension, p.Position, len(p.Offsets))
	case *packet.SubChunk:
		results := map[byte]int{}
		for _, e := range p.SubChunkEntries {
			results[e.Result]++
		}
		return fmt.Sprintf("%s pos=%v cache=%v results=%v", name, p.Position, p.CacheEnabled, results)
	case *packet.ClientCacheBlobStatus:
		return fmt.Sprintf("%s miss=%d hit=%d", name, len(p.MissHashes), len(p.HitHashes))
	case *packet.ClientCacheMissResponse:
		return fmt.Sprintf("%s blobs=%d", name, len(p.Blobs))
	case *packet.NetworkChunkPublisherUpdate:
		return fmt.Sprintf("%s pos=%v radius=%d saved=%d", name, p.Position, p.Radius, len(p.SavedChunks))
	case *packet.Disconnect:
		return fmt.Sprintf("%s reason=%d message=%q", name, p.Reason, p.Message)
	case *packet.PacketViolationWarning:
		return fmt.Sprintf("%s type=%d severity=%d packetID=%d context=%q", name, p.Type, p.Severity, p.PacketID, p.ViolationContext)
	case *packet.ServerBoundLoadingScreen:
		return fmt.Sprintf("%s type=%d", name, p.Type)
	case *packet.PlayStatus:
		return fmt.Sprintf("%s status=%d", name, p.Status)
	case *packet.ChunkRadiusUpdated:
		return fmt.Sprintf("%s radius=%d", name, p.ChunkRadius)
	case *packet.RequestChunkRadius:
		return fmt.Sprintf("%s radius=%d max=%d", name, p.ChunkRadius, p.MaxChunkRadius)
	}
	return name
}

// handleConn is gophertunnel's proxy example's handleConn, with logging.
func handleConn(conn *minecraft.Conn, listener *minecraft.Listener, remote string, logFile *os.File) {
	t := &trace{file: logFile, start: time.Now()}
	t.note("client %s connected (client blob cache enabled: %v)", conn.IdentityData().DisplayName, conn.ClientCacheEnabled())

	serverConn, err := minecraft.Dialer{
		IdentityData:      conn.IdentityData(),
		ClientData:        conn.ClientData(),
		EnableClientCache: conn.ClientCacheEnabled(),
	}.Dial("raknet", remote)
	if err != nil {
		log.Printf("dial %s: %v", remote, err)
		_ = listener.Disconnect(conn, err.Error())
		return
	}

	var g sync.WaitGroup
	g.Add(2)
	go func() {
		if err := conn.StartGame(serverConn.GameData()); err != nil {
			t.note("client StartGame: %v", err)
		}
		g.Done()
	}()
	go func() {
		if err := serverConn.DoSpawn(); err != nil {
			t.note("server spawn: %v", err)
		}
		g.Done()
	}()
	g.Wait()
	t.note("spawned; relaying")

	done := make(chan struct{}, 2)
	go func() {
		defer func() { done <- struct{}{} }()
		for {
			pk, err := conn.ReadPacket()
			if err != nil {
				t.note("client connection ended: %v", err)
				return
			}
			t.record("client->server", pk)
			if err := serverConn.WritePacket(pk); err != nil {
				t.note("write to server: %v", err)
				return
			}
		}
	}()
	go func() {
		defer func() { done <- struct{}{} }()
		for {
			pk, err := serverConn.ReadPacket()
			if err != nil {
				t.note("server connection ended: %v", err)
				return
			}
			t.record("server->client", pk)
			if err := conn.WritePacket(pk); err != nil {
				t.note("write to client: %v", err)
				return
			}
		}
	}()
	<-done
	_ = serverConn.Close()
	_ = listener.Disconnect(conn, "connection lost")

	t.mu.Lock()
	log.Printf("Session ended. Last packets (all of them are in %s):\n%s", logFile.Name(), strings.Join(t.recent, "\n"))
	t.mu.Unlock()
}
