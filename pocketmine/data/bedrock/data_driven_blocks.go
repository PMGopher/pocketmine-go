package bedrock

import (
	"bytes"
	_ "embed"
	"sync"

	gtnbt "github.com/sandertv/gophertunnel/minecraft/nbt"
)

// dataDrivenBlocksData is assets/data_driven_blocks.nbt: the component definitions of the vanilla
// blocks that Bedrock (since 1.26.50) defines through block components instead of hard-coding
// them in the client (the new wool/concrete slabs and stairs, red shrub, shelf mushroom). The
// client only completes its block palette with these blocks once the server sends their
// definitions in StartGame's block list, so without them their states (which are part of
// canonical_block_states.nbt) wouldn't line up with the client's runtime IDs.
//
// pmmp/BedrockData has no 1.26.50 release yet, so this file (like canonical_block_states.nbt and
// required_item_list.json at this version) is vendored from df-mc/dragonfly (MIT, see
// assets/LICENSE-dragonfly), which dumps the same data from the same client version.
//
//go:embed assets/data_driven_blocks.nbt
var dataDrivenBlocksData []byte

// DataDrivenBlock is one entry of data_driven_blocks.nbt: a block name and its component NBT.
type DataDrivenBlock struct {
	Name       string         `nbt:"name"`
	Components map[string]any `nbt:"components"`
}

var (
	dataDrivenBlocksOnce sync.Once
	dataDrivenBlocks     []DataDrivenBlock
)

// DataDrivenBlocks returns the vanilla data-driven block definitions the client needs in
// StartGame's block list.
func DataDrivenBlocks() []DataDrivenBlock {
	dataDrivenBlocksOnce.Do(func() {
		var m struct {
			Blocks []DataDrivenBlock `nbt:"blocks"`
		}
		if err := gtnbt.NewDecoderWithEncoding(bytes.NewReader(dataDrivenBlocksData), gtnbt.LittleEndian).Decode(&m); err != nil {
			panic("bedrock: invalid data_driven_blocks.nbt: " + err.Error())
		}
		dataDrivenBlocks = m.Blocks
	})
	return dataDrivenBlocks
}
