package bedrock

import (
	"bytes"
	_ "embed"
	"sync"

	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// The data behind pocketmine\network\mcpe\cache\StaticPacketCache: the two packets
// PreSpawnPacketHandler sends every player right after StartGame and ItemRegistry.

// entityIdentifiersData is assets/entity_identifiers.nbt: the network NBT
// StaticPacketCache::getAvailableActorIdentifiers sends as-is. pmmp/BedrockData's file is for
// 1.26.30 (136 entries with 1.26.30 runtime IDs); a 1.26.51 client that plays on Dragonfly left this
// server right after loading while it sent that list, so this is the list Dragonfly sends to
// 1.26.50 clients (MIT, see assets/LICENSE-dragonfly), captured from the wire.
//
//go:embed assets/entity_identifiers.nbt
var entityIdentifiersData []byte

// biomeDefinitionsData is assets/biome_definitions.bin: an encoded BiomeDefinitionList packet
// payload for Bedrock 1.26.50. PHP's StaticPacketCache::getBiomeDefs builds it from BedrockData's
// biome_definitions.json, which has no 1.26.50 release yet; this is the packet df-mc/dragonfly (MIT,
// see assets/LICENSE-dragonfly) sends to 1.26.50 clients, captured from the wire.
//
//go:embed assets/biome_definitions.bin
var biomeDefinitionsData []byte

// creativeContentData and craftingDataData are encoded CreativeContent and CraftingData packet
// payloads for Bedrock 1.26.50: what InventoryManager::syncCreative (CreativeInventoryCache) and
// CraftingDataCache send. PHP builds them from BedrockData's creative and recipe files, which have
// no 1.26.50 release yet, so these are the packets Dragonfly sends (MIT, see
// assets/LICENSE-dragonfly), captured from the wire. They use the item network IDs of
// required_item_list.json, which comes from the same source.
//
//go:embed assets/creative_content.bin
var creativeContentData []byte

//go:embed assets/crafting_data.bin
var craftingDataData []byte

var (
	biomeDefinitionsOnce sync.Once
	biomeDefinitions     *packet.BiomeDefinitionList

	creativeContentOnce sync.Once
	creativeContent     *packet.CreativeContent

	craftingDataOnce sync.Once
	craftingData     *packet.CraftingData
)

// CreativeContent is the creative inventory packet (InventoryManager::syncCreative).
func CreativeContent() *packet.CreativeContent {
	creativeContentOnce.Do(func() {
		pk := &packet.CreativeContent{}
		pk.Marshal(protocol.NewReader(bytes.NewReader(creativeContentData), 0, false))
		creativeContent = pk
	})
	return creativeContent
}

// CraftingData is the recipe packet (CraftingDataCache::getCache).
func CraftingData() *packet.CraftingData {
	craftingDataOnce.Do(func() {
		pk := &packet.CraftingData{}
		pk.Marshal(protocol.NewReader(bytes.NewReader(craftingDataData), 0, false))
		craftingData = pk
	})
	return craftingData
}

// AvailableActorIdentifiers is a port of StaticPacketCache::getAvailableActorIdentifiers.
func AvailableActorIdentifiers() *packet.AvailableActorIdentifiers {
	return &packet.AvailableActorIdentifiers{SerialisedEntityIdentifiers: entityIdentifiersData}
}

// BiomeDefinitionList is a port of StaticPacketCache::getBiomeDefs.
func BiomeDefinitionList() *packet.BiomeDefinitionList {
	biomeDefinitionsOnce.Do(func() {
		pk := &packet.BiomeDefinitionList{}
		pk.Marshal(protocol.NewReader(bytes.NewReader(biomeDefinitionsData), 0, false))
		biomeDefinitions = pk
	})
	return biomeDefinitions
}
