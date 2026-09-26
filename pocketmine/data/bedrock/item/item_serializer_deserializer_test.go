package bedrockitem_test

import (
	"testing"

	"pocketmine-go/pocketmine/block"
	. "pocketmine-go/pocketmine/data/bedrock/item"
	"pocketmine-go/pocketmine/item"
	worldio "pocketmine-go/pocketmine/world/format/io"
)

// These are PocketMine-MP's ItemSerializerDeserializerTest.

func roundTrip(t *testing.T, what string, it item.Item, serializer *ItemSerializer, deserializer *ItemDeserializer) {
	t.Helper()
	data, err := serializer.SerializeType(it)
	if err != nil {
		t.Errorf("%s: serialize: %v", what, err)
		return
	}
	back, err := deserializer.DeserializeType(data)
	if err != nil {
		t.Errorf("%s: deserialize %s: %v", what, data.Name, err)
		return
	}
	if !it.EqualsExact(back) {
		t.Errorf("%s: round trip through %s gave %s", what, data.Name, back)
	}
}

func TestAllVanillaItemsSerializableAndDeserializable(t *testing.T) {
	serializer, deserializer := worldio.GetItemSerializer(), worldio.GetItemDeserializer()
	for _, name := range item.GetVanillaItemNames() {
		it := item.VanillaItem(name)
		if it.IsNull() {
			continue
		}
		roundTrip(t, name, it, serializer, deserializer)
	}
}

func TestAllVanillaBlocksSerializableAndDeserializable(t *testing.T) {
	serializer, deserializer := worldio.GetItemSerializer(), worldio.GetItemDeserializer()
	for _, state := range block.GetRuntimeBlockStateRegistry().GetAllKnownStates() {
		blockItem, err := state.(interface{ AsItem() (block.Item, error) }).AsItem()
		if err != nil {
			t.Fatal(err)
		}
		it := blockItem.(item.Item)
		if it.IsNull() {
			continue
		}
		roundTrip(t, state.GetName(), it, serializer, deserializer)
	}
}
