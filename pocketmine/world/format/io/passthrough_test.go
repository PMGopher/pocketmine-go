package io

import (
	"reflect"
	"testing"

	"pocketmine-go/pocketmine/block"
	"pocketmine-go/pocketmine/data/bedrock"
	blockconvert "pocketmine-go/pocketmine/data/bedrock/block/convert"
)

// Vanilla states with no block here (moss_block, pointed_dripstone, ...) keep their state data
// on disk and their runtime ID on the network, instead of becoming "update!".
func TestVanillaStatesWithoutBlockPassThrough(t *testing.T) {
	deserializer := GetBlockStateDeserializer()
	serializer := GetBlockStateSerializer()

	count := 0
	for runtimeID, stateData := range bedrock.BlockStates() {
		if _, err := deserializer.Deserialize(stateData); err == nil {
			continue
		}
		count++
		stateID, err := deserializer.DeserializeOrPassthrough(stateData)
		if err != nil {
			t.Fatalf("%s %v: %v", stateData.Name, stateData.States, err)
		}
		if !blockconvert.IsPassthroughState(stateID) || int(int32(stateID)) != stateID {
			t.Fatalf("%s: state ID %d isn't a pass-through state", stateData.Name, stateID)
		}
		if !block.GetRuntimeBlockStateRegistry().HasStateId(stateID) {
			t.Fatalf("%s: state ID %d isn't registered", stateData.Name, stateID)
		}
		back, err := serializer.Serialize(stateID)
		if err != nil {
			t.Fatal(err)
		}
		if back.Name != stateData.Name || !reflect.DeepEqual(back.States, stateData.States) {
			t.Fatalf("runtime ID %d: got %s %v back, want %s %v", runtimeID, back.Name, back.States, stateData.Name, stateData.States)
		}
	}
	t.Logf("%d pass-through states", count)

	moss, ok := bedrock.RuntimeIDFor("minecraft:moss_block", map[string]any{})
	if !ok {
		t.Fatal("moss_block isn't in the palette")
	}
	if _, err := deserializer.DeserializeOrPassthrough(bedrock.BlockStates()[moss]); err != nil {
		t.Fatalf("moss_block: %v", err)
	}

	if _, err := deserializer.DeserializeOrPassthrough(bedrock.BlockStateData{Name: "minecraft:not_a_block", States: map[string]any{}}); err == nil {
		t.Fatal("a state that isn't in the palette must still fail")
	}
}
