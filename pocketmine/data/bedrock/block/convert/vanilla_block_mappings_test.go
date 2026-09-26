package blockconvert

import (
	"testing"

	"pocketmine-go/pocketmine/block"
	blockutils "pocketmine-go/pocketmine/block/utils"
	"pocketmine-go/pocketmine/data/bedrock"
)

func newVanillaRegistrar() *BlockSerializerDeserializerRegistrar {
	reg := NewBlockSerializerDeserializerRegistrar(NewBlockStateToObjectDeserializer(), NewBlockObjectToStateSerializer())
	InitVanillaBlockMappings(reg)
	return reg
}

// TestAllBlockStatesRoundTrip checks, for every known block state, that it serializes to a state
// in the vendored 1.26.50 palette and deserializes back to the same state
// (PocketMine-MP's BlockSerializerDeserializerTest).
func TestAllBlockStatesRoundTrip(t *testing.T) {
	reg := newVanillaRegistrar()
	failures := 0
	for _, state := range block.GetRuntimeBlockStateRegistry().GetAllKnownStates() {
		state = state.Clone() //the fix-ups below may modify it
		data, err := reg.Serializer.SerializeBlock(state)
		if err != nil {
			failures++
			if failures <= 40 {
				t.Errorf("%s (%T): serialize: %v", state, state, err)
			}
			continue
		}
		if _, ok := bedrock.RuntimeIDFor(data.Name, data.States); !ok {
			failures++
			if failures <= 40 {
				t.Errorf("%s: serialized to %s %v, which isn't in the palette", state, data.Name, data.States)
			}
			continue
		}
		back, err := reg.Deserializer.DeserializeBlock(data)
		if err != nil {
			failures++
			if failures <= 40 {
				t.Errorf("%s: deserialize %s %v: %v", state, data.Name, data.States, err)
			}
			continue
		}
		if skipRoundTripCheck(state) {
			continue
		}
		fixUpRoundTrip(state, back)
		if back.GetStateId() != state.GetStateId() {
			failures++
			if failures <= 40 {
				t.Errorf("%s: round trip through %s %v gave %s", state, data.Name, data.States, back)
			}
		}
	}
	if failures > 0 {
		t.Errorf("%d failures", failures)
	}
}

// skipRoundTripCheck and fixUpRoundTrip are BlockSerializerDeserializerTest's workarounds:
// some blocks pretend to be something else in the blockstate (the variant switching is done via
// block entity data), and some properties aren't stored in the blockstate (but in the block entity
// NBT) although they're part of the internal state.
func skipRoundTripCheck(state block.Behavior) bool {
	switch state.GetTypeId() {
	case block.POTION_CAULDRON, block.OMINOUS_BANNER, block.OMINOUS_WALL_BANNER:
		return true
	}
	return false
}

func fixUpRoundTrip(original, back block.Behavior) {
	type coloredBlock interface {
		GetColor() blockutils.DyeColor
		SetColor(color blockutils.DyeColor)
	}
	switch o := original.(type) {
	case *block.FloorBanner, *block.WallBanner, *block.Bed:
		back.(coloredBlock).SetColor(o.(coloredBlock).GetColor())
	case *block.MobHead:
		back.(*block.MobHead).SetMobHeadType(o.GetMobHeadType())
	case *block.CaveVines:
		if !o.HasBerries() {
			back.(*block.CaveVines).SetHead(o.IsHead())
		}
	case *block.Farmland:
		o.SetWaterPositionIndex(back.(*block.Farmland).GetWaterPositionIndex())
	}
}
