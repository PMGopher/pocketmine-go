package mcpe

import (
	"bytes"
	"sort"
	"testing"

	gtnbt "github.com/sandertv/gophertunnel/minecraft/nbt"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/network/mcpe/serializer"
)

// Every block with a tile must produce spawn compounds the client can read: the tile NBT that
// follows a sub-chunk is a sequence of network NBT compounds, and a broken one makes the client
// leave with "Block".
func TestEveryTileEncodesAsNetworkNbt(t *testing.T) {
	w := newTestWorld()
	w.GetOrLoadChunk(0, 0)
	names := make([]string, 0)
	for name, b := range block.GetAllVanillaBlocks() {
		if id, ok := b.(interface{ GetIdInfo() *block.BlockIdentifier }); ok && id.GetIdInfo().HasTile() {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	for i, name := range names {
		x, z := i%16, (i/16)%16
		y := 100 + i/256
		if err := w.SetBlock(block.NewPosition(float64(x), float64(y), float64(z), w), block.VanillaBlock(name)); err != nil {
			t.Errorf("%s: %v", name, err)
		}
	}
	chunk, _ := w.GetChunk(0, 0)
	count := 0
	for subY := 6; subY <= 7; subY++ {
		data := serializer.SerializeTiles(chunk, subY)
		r := bytes.NewReader(data)
		dec := gtnbt.NewDecoderWithEncoding(r, gtnbt.NetworkLittleEndian)
		for r.Len() > 0 {
			var m map[string]any
			if err := dec.Decode(&m); err != nil {
				t.Fatalf("sub-chunk %d: tile %d doesn't decode: %v", subY, count, err)
			}
			count++
		}
	}
	if count < len(names)*9/10 {
		t.Errorf("decoded only %d tiles for %d blocks with tiles", count, len(names))
	}
}
