package world

import (
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/world/format"
)

// recordingListener is a chunk listener that is also a packet viewer, like a Player.
type recordingListener struct {
	packets []packet.Packet
	changed [][2]int
}

func (l *recordingListener) SendPacket(pk packet.Packet) { l.packets = append(l.packets, pk) }
func (l *recordingListener) OnChunkChanged(chunkX, chunkZ int, chunk *format.Chunk) {
	l.changed = append(l.changed, [2]int{chunkX, chunkZ})
}
func (l *recordingListener) OnChunkLoaded(chunkX, chunkZ int, chunk *format.Chunk)    {}
func (l *recordingListener) OnChunkUnloaded(chunkX, chunkZ int, chunk *format.Chunk)  {}
func (l *recordingListener) OnChunkPopulated(chunkX, chunkZ int, chunk *format.Chunk) {}
func (l *recordingListener) OnBlockChanged(pos math.Vector3)                          {}

func TestChangedBlocksAreSentToChunkPlayersAtTheEndOfTheTick(t *testing.T) {
	w := newTestWorld()
	w.GetOrLoadChunk(0, 0)
	listener := &recordingListener{}
	w.RegisterChunkListener(listener, 0, 0)

	if err := w.SetBlockAt(1, 10, 2, block.VanillaStone()); err != nil {
		t.Fatal(err)
	}
	if len(listener.packets) != 0 {
		t.Fatal("block update sent before the end of the tick")
	}
	w.DoTick(1)

	var updates []*packet.UpdateBlock
	for _, pk := range listener.packets {
		if u, ok := pk.(*packet.UpdateBlock); ok {
			updates = append(updates, u)
		}
	}
	if len(updates) != 1 {
		t.Fatalf("got %d UpdateBlock packets, want 1", len(updates))
	}
	want := uint32(w.Translator().InternalIDToNetworkID(block.VanillaStone().GetStateId()))
	if updates[0].Position != (protocol.BlockPos{1, 10, 2}) || updates[0].NewBlockRuntimeID != want {
		t.Errorf("UpdateBlock = %+v, want stone at 1,10,2", updates[0])
	}

	listener.packets = nil
	w.DoTick(2)
	if len(listener.packets) != 0 {
		t.Error("changed blocks were sent again on the next tick")
	}
}

func TestSetBlockUpdateFalseSkipsNeighbourUpdates(t *testing.T) {
	w := newTestWorld()
	w.GetOrLoadChunk(0, 0)
	pos := block.NewPosition(1, 10, 1, w)

	if err := w.SetBlockUpdate(pos, block.VanillaStone(), false); err != nil {
		t.Fatal(err)
	}
	if len(w.neighbourUpdateQueue) != 0 {
		t.Errorf("setBlock(..., false) queued %d neighbour updates, want none", len(w.neighbourUpdateQueue))
	}
	if err := w.SetBlock(pos, block.VanillaDirt()); err != nil {
		t.Fatal(err)
	}
	if len(w.neighbourUpdateQueue) != 7 {
		t.Errorf("setBlock queued %d neighbour updates, want 7 (the block and its 6 sides)", len(w.neighbourUpdateQueue))
	}
}
