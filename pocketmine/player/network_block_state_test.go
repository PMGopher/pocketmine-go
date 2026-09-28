package player

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/data/bedrock"
	"pocketmine-go/pocketmine/math"
	"pocketmine-go/pocketmine/network/mcpe/serializer"
	"pocketmine-go/pocketmine/world/format"
)

// connectedStates returns which 1.26.50 connection flags the network state networkID has set.
func connectedStates(networkID int32) map[string]bool {
	result := map[string]bool{}
	for k, v := range bedrock.BlockStates()[networkID].States {
		if b, ok := v.(uint8); ok && b == 1 {
			result[k] = true
		}
	}
	return result
}

// subChunkPalette decodes the palette of the first layer of a serialised sub-chunk.
func subChunkPalette(t *testing.T, data []byte) []int32 {
	t.Helper()
	r := bytes.NewReader(data[3:]) // version, layer count, Y index
	header, _ := r.ReadByte()
	bitsPerBlock := int(header >> 1)
	size, err := format.GetExpectedWordArraySize(bitsPerBlock)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Seek(int64(size), 1); err != nil {
		t.Fatal(err)
	}
	count := int64(1)
	if bitsPerBlock != 0 {
		count, _ = binary.ReadVarint(r)
	}
	palette := make([]int32, count)
	for i := range palette {
		v, _ := binary.ReadVarint(r)
		palette[i] = int32(v)
	}
	return palette
}

func TestFenceConnectionsReachTheClient(t *testing.T) {
	p := newTestPlayer(t, 1, math.NewVector3(0.5, 70, 0.5))
	w := p.GetWorld()
	setBlockAt(t, p, 0, 70, 2, block.VanillaBlock("oak_fence"))
	setBlockAt(t, p, 1, 70, 2, block.VanillaBlock("oak_fence"))

	translator := w.Translator()
	got := connectedStates(translator.NetworkIDForBlock(w.GetBlockAt(0, 70, 2)))
	if !got["minecraft:connection_east"] || got["minecraft:connection_west"] {
		t.Errorf("fence at x=0 connections = %v, want only east", got)
	}

	// Placing the second fence must re-send the first one (its connections changed).
	var updated []int32
	for _, pk := range w.CreateBlockUpdatePackets([]math.Vector3{math.NewVector3(1, 70, 2)}) {
		if u, ok := pk.(*packet.UpdateBlock); ok {
			updated = append(updated, u.Position.X())
		}
	}
	if len(updated) != 2 {
		t.Errorf("updated block x positions = %v, want the changed fence and its neighbour", updated)
	}

	// Sub-chunk data carries the connected states too.
	chunk, _ := w.GetChunk(0, 0)
	data := serializer.SerializeSubChunk(chunk.GetSubChunk(70>>4), 70>>4, translator, func(x, y, z int) block.Behavior {
		return w.GetBlockAt(x, (70>>4)<<4|y, z)
	})
	found := false
	for _, id := range subChunkPalette(t, data) {
		if s := connectedStates(id); s["minecraft:connection_east"] && bedrock.BlockStates()[id].Name == "minecraft:oak_fence" {
			found = true
		}
	}
	if !found {
		t.Error("the sub-chunk palette has no east-connected oak fence")
	}
}

func TestStairCornerReachesTheClient(t *testing.T) {
	p := newTestPlayer(t, 1, math.NewVector3(0.5, 70, 0.5))
	w := p.GetWorld()
	facing := func(f math.Facing) block.Behavior {
		s := block.VanillaBlock("oak_stairs")
		s.(interface{ SetFacing(math.Facing) }).SetFacing(f)
		return s
	}
	setBlockAt(t, p, 0, 70, 2, facing(math.North))
	setBlockAt(t, p, 0, 70, 1, facing(math.East))
	id := w.Translator().NetworkIDForBlock(w.GetBlockAt(0, 70, 2))
	if corner := bedrock.BlockStates()[id].States["minecraft:corner"]; corner == "none" {
		t.Errorf("corner of stairs next to a perpendicular stair = %v", corner)
	}
}
